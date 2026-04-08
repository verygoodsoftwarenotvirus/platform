package analyticscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

// ProvideEventReporter provides an analytics.EventReporter from a config.
func ProvideEventReporter(ctx context.Context, cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (analytics.EventReporter, error) {
	return cfg.ProvideCollector(ctx, logger, tracerProvider, metricsProvider)
}
