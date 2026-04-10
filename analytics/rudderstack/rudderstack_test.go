package rudderstack

import (
	"errors"
	"testing"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

func TestNewRudderstackEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, nil, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)
	})

	T.Run("with empty API key", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			APIKey:       "",
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)
	})

	T.Run("with empty DataPlane URL", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: "",
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)
	})

	T.Run("with error creating event counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_events", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		collector, err := NewRudderstackEventReporter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_events", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		collector, err := NewRudderstackEventReporter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestRudderstackEventReporter_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		collector.Close()
	})
}

func TestRudderstackEventReporter_AddUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.AddUser(ctx, exampleUserID, properties))
	})
}

func TestRudderstackEventReporter_EventOccurred(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.EventOccurred(ctx, t.Name(), exampleUserID, properties))
	})
}

func TestRudderstackEventReporter_EventOccurredAnonymous(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		exampleAnonymousID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		collector, err := NewRudderstackEventReporter(logger, tracing.NewNoopTracerProvider(), nil, cfg, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.EventOccurredAnonymous(ctx, t.Name(), exampleAnonymousID, properties))
	})
}
