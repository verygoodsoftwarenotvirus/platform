package tracingcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
)

// ProvideTracerProvider provides a TracerProvider from config.
func ProvideTracerProvider(ctx context.Context, c *Config, l logging.Logger) (traceProvider tracing.TracerProvider, err error) {
	return c.ProvideTracerProvider(ctx, l)
}
