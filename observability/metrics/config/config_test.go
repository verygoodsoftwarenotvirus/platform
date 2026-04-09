package metricscfg

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/otelgrpc"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ProvideMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		metricsProvider, err := cfg.ProvideMetricsProvider(t.Context(), logging.NewNoopLogger())

		assert.NoError(t, err)
		assert.NotNil(t, metricsProvider)
	})

	T.Run("enabled with otel provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled:     true,
			Provider:    ProviderOtel,
			ServiceName: t.Name(),
			Otel: &otelgrpc.Config{
				CollectorEndpoint:  "localhost:4317",
				CollectionInterval: 30 * time.Second,
				Insecure:           true,
			},
		}

		metricsProvider, err := cfg.ProvideMetricsProvider(t.Context(), logging.NewNoopLogger())

		assert.NoError(t, err)
		assert.NotNil(t, metricsProvider)
	})

	T.Run("enabled with unknown provider falls back to noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled:  true,
			Provider: "unknown",
		}

		metricsProvider, err := cfg.ProvideMetricsProvider(t.Context(), logging.NewNoopLogger())

		assert.NoError(t, err)
		assert.NotNil(t, metricsProvider)
	})

	T.Run("not enabled returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled: false,
		}

		metricsProvider, err := cfg.ProvideMetricsProvider(t.Context(), logging.NewNoopLogger())

		assert.NoError(t, err)
		assert.NotNil(t, metricsProvider)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled:  true,
			Provider: ProviderOtel,
			Otel: &otelgrpc.Config{
				CollectorEndpoint:  t.Name(),
				CollectionInterval: 1,
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("disabled is valid", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled: false,
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("enabled with invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled:  true,
			Provider: "bogus",
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("enabled with otel provider but nil otel config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Enabled:  true,
			Provider: ProviderOtel,
			Otel:     nil,
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestProvideMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		metricsProvider, err := ProvideMetricsProvider(t.Context(), logging.NewNoopLogger(), cfg)

		assert.NoError(t, err)
		assert.NotNil(t, metricsProvider)
	})
}
