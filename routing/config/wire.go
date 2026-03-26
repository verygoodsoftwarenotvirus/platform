package routingcfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v3/routing"

	"github.com/google/wire"
)

var (
	// RoutingConfigProviders are what we provide to the dependency injector.
	RoutingConfigProviders = wire.NewSet(
		// ProvideRouterViaConfig,
		ProvideRouteParamManager,
	)
)

func ProvideRouterViaConfig(
	cfg *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricProvider metrics.Provider,
) (routing.Router, error) {
	return cfg.ProvideRouter(logger, tracerProvider, metricProvider)
}
