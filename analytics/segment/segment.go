package segment

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/errors"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	segment "github.com/segmentio/analytics-go/v3"
)

const (
	name = "segment_event_reporter"
)

var (
	// ErrEmptyAPIToken indicates an empty API token was provided.
	ErrEmptyAPIToken = platformerrors.New("empty Segment API token")
)

type (
	// EventReporter is a Segment-backed EventReporter.
	EventReporter struct {
		tracer         tracing.Tracer
		logger         logging.Logger
		client         segment.Client
		circuitBreaker circuitbreaking.CircuitBreaker
	}
)

// NewSegmentEventReporter returns a new Segment-backed EventReporter.
func NewSegmentEventReporter(logger logging.Logger, tracerProvider tracing.TracerProvider, apiKey string, circuitBreaker circuitbreaking.CircuitBreaker) (analytics.EventReporter, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIToken
	}

	c := &EventReporter{
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(name)),
		logger:         logging.EnsureLogger(logger).WithName(name),
		client:         segment.New(apiKey),
		circuitBreaker: circuitBreaker,
	}

	return c, nil
}

// Close wraps the internal client's Close method.
func (c *EventReporter) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.Error("closing connection", err)
	}
}

// AddUser upsert's a user's identity.
func (c *EventReporter) AddUser(ctx context.Context, userID string, properties map[string]any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if c.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	t := segment.NewTraits()
	for k, v := range properties {
		t.Set(k, v)
	}

	i := segment.NewIntegrations().EnableAll()

	err := c.client.Enqueue(segment.Identify{
		UserId:       userID,
		Traits:       t,
		Integrations: i,
	})
	if err != nil {
		c.circuitBreaker.Failed()
		return err
	}

	c.circuitBreaker.Succeeded()
	return nil
}

// EventOccurred associates events with a user.
func (c *EventReporter) EventOccurred(ctx context.Context, event, userID string, properties map[string]any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if c.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	p := segment.NewProperties()
	for k, v := range properties {
		p.Set(k, v)
	}

	if err := c.client.Enqueue(segment.Track{
		Event:        event,
		UserId:       userID,
		Properties:   p,
		Integrations: segment.NewIntegrations().EnableAll(),
	}); err != nil {
		c.circuitBreaker.Failed()
		return err
	}

	c.circuitBreaker.Succeeded()
	return nil
}

// EventOccurredAnonymous records an event for an anonymous user.
func (c *EventReporter) EventOccurredAnonymous(ctx context.Context, event, anonymousID string, properties map[string]any) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if c.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	p := segment.NewProperties()
	for k, v := range properties {
		p.Set(k, v)
	}

	if err := c.client.Enqueue(segment.Track{
		Event:        event,
		AnonymousId:  anonymousID,
		Properties:   p,
		Integrations: segment.NewIntegrations().EnableAll(),
	}); err != nil {
		c.circuitBreaker.Failed()
		return err
	}

	c.circuitBreaker.Succeeded()
	return nil
}
