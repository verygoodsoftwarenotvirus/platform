package ably

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	ablyrest "github.com/ably/ably-go/ably"
)

const o11yName = "async_notifications_ably"

var (
	_ asyncnotifications.AsyncNotifier = (*Notifier)(nil)

	ErrNilConfig = errors.New("ably config is nil")
)

// ChannelPublisher abstracts Ably channel publishing for testability.
type ChannelPublisher interface {
	Publish(ctx context.Context, channel, name string, data any) error
}

// ablyChannelPublisher is the real implementation wrapping the Ably REST client.
type ablyChannelPublisher struct {
	client *ablyrest.REST
}

func (a *ablyChannelPublisher) Publish(ctx context.Context, channel, name string, data any) error {
	return a.client.Channels.Get(channel).Publish(ctx, name, data)
}

// Notifier is an Ably-backed AsyncNotifier.
type Notifier struct {
	logger    logging.Logger
	tracer    tracing.Tracer
	publisher ChannelPublisher
}

// NewNotifier creates a new Ably-backed AsyncNotifier.
func NewNotifier(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (*Notifier, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	client, err := ablyrest.NewREST(ablyrest.WithKey(cfg.APIKey))
	if err != nil {
		return nil, errors.Wrap(err, "creating ably client")
	}

	return &Notifier{
		logger:    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		publisher: &ablyChannelPublisher{client: client},
	}, nil
}

// Publish sends an event to the given Ably channel.
func (n *Notifier) Publish(ctx context.Context, channel string, event *asyncnotifications.Event) error {
	_, span := n.tracer.StartSpan(ctx)
	defer span.End()

	if err := n.publisher.Publish(ctx, channel, event.Type, event.Data); err != nil {
		return observability.PrepareAndLogError(err, n.logger, span, "publishing to ably channel")
	}

	return nil
}

// Close is a no-op for the Ably notifier (REST client, no persistent connection).
func (n *Notifier) Close() error {
	return nil
}
