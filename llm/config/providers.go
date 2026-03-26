package llmcfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/llm"
)

// ProvideLLMProvider provides an LLM provider from config.
func ProvideLLMProvider(c *Config) (llm.Provider, error) {
	return c.ProvideLLMProvider(context.Background())
}
