package analyticscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/shoenig/test/must"
)

func TestProvideCollector(T *testing.T) {
	T.Parallel()

	T.Run("noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}
		logger := logging.NewNoopLogger()

		actual, err := ProvideEventReporter(ctx, cfg, logger, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		must.NoError(t, err)
		must.NotNil(t, actual)
	})

	T.Run("with segment", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderSegment,
				Segment: &segment.Config{
					APIToken: t.Name(),
				},
			},
		}
		logger := logging.NewNoopLogger()

		actual, err := ProvideEventReporter(ctx, cfg, logger, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		must.NoError(t, err)
		must.NotNil(t, actual)
	})
}
