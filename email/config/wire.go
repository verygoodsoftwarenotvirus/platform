package emailcfg

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v3/email"
	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/google/wire"
)

var (
	// Providers are what we provide to dependency injection.
	Providers = wire.NewSet(
		ProvideEmailer,
	)
)

// ProvideEmailer provides an email.Emailer from a config.
func ProvideEmailer(ctx context.Context, cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider, client *http.Client) (email.Emailer, error) {
	circuitBreaker, err := cfg.CircuitBreaker.ProvideCircuitBreaker(ctx, logger, metricsProvider)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize email circuit breaker")
	}

	return cfg.ProvideEmailer(logger, tracerProvider, client, circuitBreaker)
}
