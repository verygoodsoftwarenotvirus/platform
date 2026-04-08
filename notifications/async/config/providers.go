package asynccfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v5/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

// ProvideAsyncNotifierFromConfig provides an AsyncNotifier from a config.
func ProvideAsyncNotifierFromConfig(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (async.AsyncNotifier, error) {
	return cfg.ProvideAsyncNotifier(logger, tracerProvider, metricsProvider)
}
