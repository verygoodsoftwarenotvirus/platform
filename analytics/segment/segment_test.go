package segment

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestNewSegmentEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		collector, err := NewSegmentEventReporter(logger, tracing.NewNoopTracerProvider(), t.Name(), circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)
	})

	T.Run("with empty API key", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		collector, err := NewSegmentEventReporter(logger, tracing.NewNoopTracerProvider(), "", circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, collector)
	})
}

func TestSegmentEventReporter_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		collector, err := NewSegmentEventReporter(logger, tracing.NewNoopTracerProvider(), t.Name(), circuitbreaking.NewNoopCircuitBreaker())
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
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		collector, err := NewSegmentEventReporter(logger, tracing.NewNoopTracerProvider(), t.Name(), circuitbreaking.NewNoopCircuitBreaker())
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
		exampleUserID := identifiers.New()
		properties := map[string]any{
			"test.name": t.Name(),
		}

		collector, err := NewSegmentEventReporter(logger, tracing.NewNoopTracerProvider(), t.Name(), circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, collector)

		require.NoError(t, collector.EventOccurred(ctx, t.Name(), exampleUserID, properties))
	})
}
