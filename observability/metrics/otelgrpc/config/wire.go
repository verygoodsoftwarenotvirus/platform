package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics/otelgrpc"

	"github.com/google/wire"
)

var (
	// Providers provides metrics provider construction for dependency injection.
	Providers = wire.NewSet(
		ProvideMetricsProvider,
	)
)

// ProvideMetricsProvider provides a metrics.Provider from the config.
func ProvideMetricsProvider(ctx context.Context, logger logging.Logger, cfg *Config) (metrics.Provider, error) {
	return otelgrpc.ProvideMetricsProvider(ctx, logger, cfg.ServiceName, cfg.Otel)
}
