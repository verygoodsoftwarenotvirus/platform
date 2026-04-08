package llmcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/llm"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

// ProvideLLMProvider provides an LLM provider from config.
func ProvideLLMProvider(c *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (llm.Provider, error) {
	return c.ProvideLLMProvider(context.Background(), logger, tracerProvider, metricsProvider)
}
