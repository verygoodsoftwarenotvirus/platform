package featureflagscfg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/google/wire"
)

var (
	ProvidersFeatureFlags = wire.NewSet(
		ProvideFeatureFlagManager,
	)
)

func ProvideFeatureFlagManager(ctx context.Context, c *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, httpClient *http.Client) (featureflags.FeatureFlagManager, error) {
	circuitBreaker, err := c.CircuitBreaker.ProvideCircuitBreaker(ctx, logger, metricsProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize feature flag circuit breaker: %w", err)
	}

	return c.ProvideFeatureFlagManager(logger, tracerProvider, httpClient, circuitBreaker)
}
