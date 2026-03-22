package analyticscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestProvideCollector(T *testing.T) {
	T.Parallel()

	T.Run("noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}
		logger := logging.NewNoopLogger()

		actual, err := ProvideEventReporter(ctx, cfg, logger, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, actual)
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
		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}
