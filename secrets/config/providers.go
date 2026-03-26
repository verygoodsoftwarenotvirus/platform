package secretscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/secrets"
	"github.com/verygoodsoftwarenotvirus/platform/v4/secrets/env"
)

// ProvideSecretSourceFromConfig provides a SecretSource from config.
func ProvideSecretSourceFromConfig(ctx context.Context, cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (secrets.SecretSource, error) {
	if cfg == nil {
		return env.NewEnvSecretSource(logger, tracerProvider, metricsProvider)
	}
	source, err := cfg.ProvideSecretSource(ctx, logger, tracerProvider, metricsProvider)
	if err != nil {
		return nil, errors.Wrap(err, "provide secret source")
	}
	return source, nil
}
