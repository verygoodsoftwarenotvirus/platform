package config

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/otelgrpc"
)

// ProvideMetricsProvider provides a metrics.Provider from the config.
func ProvideMetricsProvider(ctx context.Context, logger logging.Logger, cfg *Config) (metrics.Provider, error) {
	return otelgrpc.ProvideMetricsProvider(ctx, logger, cfg.ServiceName, cfg.Otel)
}
