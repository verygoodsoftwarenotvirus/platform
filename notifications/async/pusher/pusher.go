package pusher

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	pushersdk "github.com/pusher/pusher-http-go/v5"
)

const o11yName = "async_notifications_pusher"

var (
	_ async.AsyncNotifier = (*Notifier)(nil)

	ErrNilConfig = errors.New("pusher config is nil")
)

// PusherClient abstracts the Pusher SDK client for testability.
type PusherClient interface {
	Trigger(channel string, eventName string, data any) error
}

// Notifier is a Pusher-backed AsyncNotifier.
type Notifier struct {
	logger logging.Logger
	tracer tracing.Tracer
	client PusherClient
}

// NewNotifier creates a new Pusher-backed AsyncNotifier.
func NewNotifier(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (*Notifier, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	client := &pushersdk.Client{
		AppID:   cfg.AppID,
		Key:     cfg.Key,
		Secret:  cfg.Secret,
		Cluster: cfg.Cluster,
		Secure:  cfg.Secure,
	}

	return &Notifier{
		logger: logging.NewNamedLogger(logger, o11yName),
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		client: client,
	}, nil
}

// Publish sends an event to the given Pusher channel.
func (n *Notifier) Publish(ctx context.Context, channel string, event *async.Event) error {
	_, span := n.tracer.StartSpan(ctx)
	defer span.End()

	if err := n.client.Trigger(channel, event.Type, event.Data); err != nil {
		return observability.PrepareAndLogError(err, n.logger, span, "publishing to pusher channel")
	}

	return nil
}

// Close is a no-op for the Pusher notifier (stateless HTTP API).
func (n *Notifier) Close() error {
	return nil
}
