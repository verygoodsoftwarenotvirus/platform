package secretscfg

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v2/secrets"
	"github.com/verygoodsoftwarenotvirus/platform/v2/secrets/env"

	"github.com/google/wire"
)

var (
	// Providers provides secret source construction for dependency injection.
	Providers = wire.NewSet(
		ProvideSecretSourceFromConfig,
	)
)

// ProvideSecretSourceFromConfig provides a SecretSource from config.
func ProvideSecretSourceFromConfig(ctx context.Context, cfg *Config) (secrets.SecretSource, error) {
	if cfg == nil {
		return env.NewEnvSecretSource(), nil
	}
	source, err := cfg.ProvideSecretSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("provide secret source: %w", err)
	}
	return source, nil
}
