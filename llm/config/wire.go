package llmcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/llm"

	"github.com/google/wire"
)

var (
	// ProvidersLLM provides the LLM provider for Wire dependency injection.
	ProvidersLLM = wire.NewSet(
		ProvideLLMProvider,
	)
)

// ProvideLLMProvider provides an LLM provider from config.
func ProvideLLMProvider(c *Config) (llm.Provider, error) {
	return c.ProvideLLMProvider(context.Background())
}
