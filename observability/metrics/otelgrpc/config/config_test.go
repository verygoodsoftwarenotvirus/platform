package config

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/otelgrpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			CollectorEndpoint: t.Name(),
			ServiceName:       t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("missing collector endpoint", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			ServiceName: t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("missing service name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			CollectorEndpoint: t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ServiceName:       t.Name(),
			CollectorEndpoint: "localhost:4317",
			Otel: &otelgrpc.Config{
				CollectorEndpoint:  "localhost:4317",
				CollectionInterval: 30 * time.Second,
				Insecure:           true,
			},
		}

		provider, err := ProvideMetricsProvider(t.Context(), logging.NewNoopLogger(), cfg)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	T.Run("with nil otel config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ServiceName:       t.Name(),
			CollectorEndpoint: "localhost:4317",
		}

		provider, err := ProvideMetricsProvider(t.Context(), logging.NewNoopLogger(), cfg)
		assert.Nil(t, provider)
		assert.Error(t, err)
	})
}
