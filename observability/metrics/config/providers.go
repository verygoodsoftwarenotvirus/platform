package metricscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
)

// ProvideMetricsProvider provides a metrics.Provider from config.
func ProvideMetricsProvider(ctx context.Context, logger logging.Logger, c *Config) (metrics.Provider, error) {
	return c.ProvideMetricsProvider(ctx, logger)
}
