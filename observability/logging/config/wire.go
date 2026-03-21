package loggingcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

	"github.com/google/wire"
)

var (
	LogConfigProviders = wire.NewSet(
		ProvideLogger,
	)
)

func ProvideLogger(ctx context.Context, cfg *Config) (logging.Logger, error) {
	return cfg.ProvideLogger(ctx)
}
