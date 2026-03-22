package featureflagscfg

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/featureflags"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

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
		return nil, errors.Wrap(err, "failed to initialize feature flag circuit breaker")
	}

	return c.ProvideFeatureFlagManager(logger, tracerProvider, httpClient, circuitBreaker)
}
