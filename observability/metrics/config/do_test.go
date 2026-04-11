package metricscfg

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"

	"github.com/samber/do/v2"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func TestRegisterMetricsProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue[context.Context](i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, &Config{})

		RegisterMetricsProvider(i)

		mp, err := do.Invoke[metrics.Provider](i)
		must.NoError(t, err)
		test.NotNil(t, mp)
	})
}
