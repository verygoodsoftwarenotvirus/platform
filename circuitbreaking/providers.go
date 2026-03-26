package circuitbreaking

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics"
)

// ProvideCircuitBreaker provides a CircuitBreaker from config.
func ProvideCircuitBreaker(ctx context.Context, cfg *Config, logger logging.Logger, metricsProvider metrics.Provider) (CircuitBreaker, error) {
	return cfg.ProvideCircuitBreaker(ctx, logger, metricsProvider)
}
