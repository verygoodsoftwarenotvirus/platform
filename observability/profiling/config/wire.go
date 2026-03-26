package profilingcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/profiling"

	"github.com/google/wire"
)

var (
	// ProfilingConfigProviders is a Wire provider set for profiling.
	ProfilingConfigProviders = wire.NewSet(
		ProvideProfilingProviderWire,
	)
)

// ProvideProfilingProviderWire provides a profiling provider from config.
func ProvideProfilingProviderWire(ctx context.Context, logger logging.Logger, c *Config) (profiling.Provider, error) {
	return c.ProvideProfilingProvider(ctx, logger)
}
