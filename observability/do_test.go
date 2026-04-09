package observability

import (
	"testing"

	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/config"
	profilingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/profiling/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing/config"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterO11yConfigs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		i := do.New()
		do.ProvideValue(i, cfg)

		RegisterO11yConfigs(i)

		loggingConfig, err := do.Invoke[*loggingcfg.Config](i)
		require.NoError(t, err)
		assert.NotNil(t, loggingConfig)

		metricsConfig, err := do.Invoke[*metricscfg.Config](i)
		require.NoError(t, err)
		assert.NotNil(t, metricsConfig)

		tracingConfig, err := do.Invoke[*tracingcfg.Config](i)
		require.NoError(t, err)
		assert.NotNil(t, tracingConfig)

		profilingConfig, err := do.Invoke[*profilingcfg.Config](i)
		require.NoError(t, err)
		assert.NotNil(t, profilingConfig)
	})
}
