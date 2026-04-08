package embeddingscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/cohere"
	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/ollama"
	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/openai"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ProvideEmbedder_Empty(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ""}
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracerForTest("test")

		embedder, err := cfg.ProvideEmbedder(t.Context(), logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, embedder, "expected non-nil embedder (noop)")
	})
}

func TestConfig_ProvideEmbedder_OpenAI(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderOpenAI,
			OpenAI: &openai.Config{
				APIKey: "test-key",
			},
		}
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracerForTest("test")

		embedder, err := cfg.ProvideEmbedder(t.Context(), logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, embedder)
	})
}

func TestConfig_ProvideEmbedder_Ollama(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderOllama,
			Ollama:   &ollama.Config{},
		}
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracerForTest("test")

		embedder, err := cfg.ProvideEmbedder(t.Context(), logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, embedder)
	})
}

func TestConfig_ProvideEmbedder_Cohere(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderCohere,
			Cohere: &cohere.Config{
				APIKey: "test-key",
			},
		}
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracerForTest("test")

		embedder, err := cfg.ProvideEmbedder(t.Context(), logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, embedder)
	})
}
