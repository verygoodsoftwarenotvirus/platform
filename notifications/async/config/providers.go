package asynccfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
)

// ProvideAsyncNotifierFromConfig provides an AsyncNotifier from a config.
func ProvideAsyncNotifierFromConfig(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (async.AsyncNotifier, error) {
	return cfg.ProvideAsyncNotifier(logger, tracerProvider)
}
