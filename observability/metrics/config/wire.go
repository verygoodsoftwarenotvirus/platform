package metricscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"

	"github.com/google/wire"
)

var (
	// MetricsConfigProviders is a Wire provider set that provides a tracing.TracerProvider.
	MetricsConfigProviders = wire.NewSet(
		ProvideMetricsProvider,
	)
)

func ProvideMetricsProvider(ctx context.Context, logger logging.Logger, c *Config) (metrics.Provider, error) {
	return c.ProvideMetricsProvider(ctx, logger)
}
