package featureflagscfg

import (
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue(i, metrics.NewNoopMetricsProvider())
		do.ProvideValue(i, http.DefaultClient)
		do.ProvideValue(i, &Config{})

		RegisterFeatureFlagManager(i)

		ffm, err := do.Invoke[featureflags.FeatureFlagManager](i)
		require.NoError(t, err)
		assert.NotNil(t, ffm)
	})
}
