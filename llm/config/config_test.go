package llmcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/llm/openai"

	"github.com/stretchr/testify/require"
)

func TestConfig_ProvideLLMProvider_Empty(t *testing.T) {
	ctx := t.Context()
	cfg := &Config{Provider: ""}

	provider, err := cfg.ProvideLLMProvider(ctx)
	require.NoError(t, err)
	require.NotNil(t, provider, "expected non-nil provider (noop)")
}

func TestConfig_ProvideLLMProvider_OpenAI(t *testing.T) {
	ctx := t.Context()
	cfg := &Config{
		Provider: ProviderOpenAI,
		OpenAI: &openai.Config{
			APIKey: "test-key",
		},
	}

	provider, err := cfg.ProvideLLMProvider(ctx)
	require.NoError(t, err)
	require.NotNil(t, provider, "expected non-nil provider")
}
