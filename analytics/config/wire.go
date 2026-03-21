package analyticscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		ProvideEventReporter,
	)
)

// ProvideEventReporter provides an analytics.EventReporter from a config.
func ProvideEventReporter(ctx context.Context, cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (analytics.EventReporter, error) {
	return cfg.ProvideCollector(ctx, logger, tracerProvider, metricsProvider)
}
