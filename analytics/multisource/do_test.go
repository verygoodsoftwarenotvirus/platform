package multisource

import (
	"testing"

	analyticscfg "github.com/verygoodsoftwarenotvirus/platform/v5/analytics/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func TestRegisterMultiSourceEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue[metrics.Provider](i, metrics.NewNoopMetricsProvider())
		do.ProvideValue(i, map[string]*analyticscfg.SourceConfig{
			"ios": {
				Provider: analyticscfg.ProviderSegment,
				Segment:  &segment.Config{APIToken: t.Name()},
			},
		})

		RegisterMultiSourceEventReporter(i)

		reporter, err := do.Invoke[*MultiSourceEventReporter](i)
		must.NoError(t, err)
		test.NotNil(t, reporter)
	})
}
