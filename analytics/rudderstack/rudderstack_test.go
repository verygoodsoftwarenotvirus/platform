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

	"github.com/shoenig/test"
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

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				test.EqOp(t, name+"_events", counterName)
				return metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary")
			},
		}

		collector, err := NewRudderstackEventReporter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			APIKey:       t.Name(),
			DataPlaneURL: t.Name(),
		}

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch counterName {
				case name + "_events":
					return metrics.Int64CounterForTest(t, "x"), nil
				case name + "_errors":
					return metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", counterName)
				return nil, nil
			},
		}

		collector, err := NewRudderstackEventReporter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, cfg, cbnoop.NewCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)

		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
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
