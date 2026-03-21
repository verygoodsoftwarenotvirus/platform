package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

var _ Tracer = (*otelTraceWrapper)(nil)

type otelTraceWrapper struct {
	tracer trace.Tracer
}

// NewTracerForTest creates a Tracer.
func NewTracerForTest(name string) Tracer {
	return &otelTraceWrapper{
		tracer: NewNoopTracerProvider().Tracer(name),
	}
}

// NewTracer creates a Tracer.
func NewTracer(t trace.Tracer) Tracer {
	return &otelTraceWrapper{
		tracer: t,
	}
}

// StartSpan wraps tracer.Start.
func (t *otelTraceWrapper) StartSpan(ctx context.Context) (context.Context, Span) {
	return t.tracer.Start(ctx, GetCallerName())
}

// StartCustomSpan wraps tracer.Start.
func (t *otelTraceWrapper) StartCustomSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return t.tracer.Start(ctx, name, opts...)
}
