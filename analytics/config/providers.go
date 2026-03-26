package analyticscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"
)

// ProvideEventReporter provides an analytics.EventReporter from a config.
func ProvideEventReporter(ctx context.Context, cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (analytics.EventReporter, error) {
	return cfg.ProvideCollector(ctx, logger, tracerProvider, metricsProvider)
}
