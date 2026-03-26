package profilingcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/profiling"
)

// ProvideProfilingProviderWire provides a profiling provider from config.
func ProvideProfilingProviderWire(ctx context.Context, logger logging.Logger, c *Config) (profiling.Provider, error) {
	return c.ProvideProfilingProvider(ctx, logger)
}
