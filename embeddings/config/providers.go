package embeddingscfg

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings"
	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/cohere"
	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/ollama"
	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings/openai"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

// ProvideEmbedder provides an Embedder from config.
func ProvideEmbedder(ctx context.Context, c *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error) {
	switch strings.TrimSpace(strings.ToLower(c.Provider)) {
	case ProviderOpenAI:
		return openai.NewEmbedder(ctx, c.OpenAI, logger, tracer)
	case ProviderOllama:
		return ollama.NewEmbedder(ctx, c.Ollama, logger, tracer)
	case ProviderCohere:
		return cohere.NewEmbedder(ctx, c.Cohere, logger, tracer)
	default:
		return embeddings.NewNoopEmbedder(), nil
	}
}
