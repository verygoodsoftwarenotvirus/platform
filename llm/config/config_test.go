package llmcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/llm/openai"

	"github.com/stretchr/testify/require"
)

func TestConfig_ProvideLLMProvider_Empty(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: ""}

		provider, err := cfg.ProvideLLMProvider(ctx, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, provider, "expected non-nil provider (noop)")
	})
}

func TestConfig_ProvideLLMProvider_OpenAI(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderOpenAI,
			OpenAI: &openai.Config{
				APIKey: "test-key",
			},
		}

		provider, err := cfg.ProvideLLMProvider(ctx, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, provider, "expected non-nil provider")
	})
}
