package observability

import (
	"testing"

	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing/oteltrace"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Tracing: tracingcfg.Config{
				ServiceName:               t.Name(),
				SpanCollectionProbability: 1,
				Provider:                  tracingcfg.ProviderOtel,
				Otel: &oteltrace.Config{
					CollectorEndpoint: "0.0.0.0",
				},
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvidePillars(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		pillars, err := cfg.ProvidePillars(ctx)
		require.NoError(t, err)
		require.NotNil(t, pillars)
		assert.NotNil(t, pillars.Logger)
		assert.NotNil(t, pillars.TracerProvider)
		assert.NotNil(t, pillars.MetricsProvider)
		assert.NotNil(t, pillars.Profiler)
	})
}
