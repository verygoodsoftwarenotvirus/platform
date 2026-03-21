package posthog

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestNewSegmentEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{APIKey: t.Name()}

		collector, err := NewPostHogEventReporter(logger, tracing.NewNoopTracerProvider(), cfg.APIKey, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)
	})

	T.Run("with empty API key", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{}

		collector, err := NewPostHogEventReporter(logger, tracing.NewNoopTracerProvider(), cfg.APIKey, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)
	})
}

func TestSegmentEventReporter_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{APIKey: t.Name()}

		collector, err := NewPostHogEventReporter(logger, tracing.NewNoopTracerProvider(), cfg.APIKey, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		collector.Close()
	})
}

func TestSegmentEventReporter_AddUser(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{APIKey: t.Name()}
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		collector, err := NewPostHogEventReporter(logger, tracing.NewNoopTracerProvider(), cfg.APIKey, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.AddUser(ctx, exampleUserID, properties))
	})
}

func TestSegmentEventReporter_EventOccurred(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{APIKey: t.Name()}
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		collector, err := NewPostHogEventReporter(logger, tracing.NewNoopTracerProvider(), cfg.APIKey, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.EventOccurred(ctx, t.Name(), exampleUserID, properties))
	})
}
