package embeddingscfg

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v4/embeddings"
	"github.com/verygoodsoftwarenotvirus/platform/v4/embeddings/cohere"
	"github.com/verygoodsoftwarenotvirus/platform/v4/embeddings/ollama"
	"github.com/verygoodsoftwarenotvirus/platform/v4/embeddings/openai"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
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
