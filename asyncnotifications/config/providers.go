package asyncnotificationscfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"
)

// ProvideAsyncNotifierFromConfig provides an AsyncNotifier from a config.
func ProvideAsyncNotifierFromConfig(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (asyncnotifications.AsyncNotifier, error) {
	return cfg.ProvideAsyncNotifier(logger, tracerProvider)
}
