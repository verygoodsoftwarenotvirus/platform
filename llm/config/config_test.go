package llmcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/llm/openai"

	"github.com/stretchr/testify/require"
)

func TestConfig_ProvideLLMProvider_Empty(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ctx := t.Context()
		cfg := &Config{Provider: ""}

		provider, err := cfg.ProvideLLMProvider(ctx)
		require.NoError(t, err)
		require.NotNil(t, provider, "expected non-nil provider (noop)")
	})
}

func TestConfig_ProvideLLMProvider_OpenAI(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
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
	})
}
