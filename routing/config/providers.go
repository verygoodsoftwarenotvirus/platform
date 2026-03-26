package routingcfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/routing"
)

// ProvideRouterViaConfig provides a Router from config.
func ProvideRouterViaConfig(
	cfg *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricProvider metrics.Provider,
) (routing.Router, error) {
	return cfg.ProvideRouter(logger, tracerProvider, metricProvider)
}
