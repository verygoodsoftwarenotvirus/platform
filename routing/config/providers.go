package routingcfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/routing"
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
