package websocket

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/eventstream"
	eswebsocket "github.com/verygoodsoftwarenotvirus/platform/v3/eventstream/websocket"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"
)

const o11yName = "async_notifications_websocket"

var (
	_ asyncnotifications.AsyncNotifier      = (*Notifier)(nil)
	_ asyncnotifications.ConnectionAcceptor = (*Notifier)(nil)

	ErrNilConfig = errors.New("websocket async notifier config is nil")
)

// Notifier is a WebSocket-backed AsyncNotifier that manages direct client connections.
type Notifier struct {
	logger   logging.Logger
	tracer   tracing.Tracer
	upgrader *eswebsocket.Upgrader
	manager  *eventstream.StreamManager[eventstream.EventStream]
}

// NewNotifier creates a new WebSocket-backed AsyncNotifier.
func NewNotifier(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (*Notifier, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	wsCfg := &eswebsocket.Config{
		HeartbeatInterval: cfg.HeartbeatInterval,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
	}

	return &Notifier{
		logger:   logging.EnsureLogger(logger).WithName(o11yName),
		tracer:   tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		upgrader: eswebsocket.NewUpgrader(tracerProvider, wsCfg),
		manager:  eventstream.NewStreamManager[eventstream.EventStream](tracerProvider, logger),
	}, nil
}

// Publish sends an event to all connected clients on the given channel.
func (n *Notifier) Publish(ctx context.Context, channel string, event *asyncnotifications.Event) error {
	_, span := n.tracer.StartSpan(ctx)
	defer span.End()

	esEvent := &eventstream.Event{
		Type:    event.Type,
		Payload: event.Data,
	}

	n.manager.BroadcastToGroup(ctx, channel, esEvent)

	return nil
}

// AcceptConnection upgrades the HTTP connection to a WebSocket and registers it
// under the given channel and memberID.
func (n *Notifier) AcceptConnection(w http.ResponseWriter, r *http.Request, channel, memberID string) error {
	ctx := r.Context()
	_, span := n.tracer.StartSpan(ctx)
	defer span.End()

	stream, err := n.upgrader.UpgradeToEventStream(w, r)
	if err != nil {
		return observability.PrepareAndLogError(err, n.logger, span, "upgrading websocket connection")
	}

	n.manager.Add(ctx, channel, memberID, stream)

	go func(removeCtx context.Context) {
		<-stream.Done()
		n.manager.Remove(removeCtx, channel, memberID)
	}(ctx)

	return nil
}

// Close releases resources held by the notifier.
func (n *Notifier) Close() error {
	return nil
}
