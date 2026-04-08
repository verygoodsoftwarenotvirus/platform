package embeddingscfg

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterEmbedder registers an embeddings.Embedder with the injector.
func RegisterEmbedder(i do.Injector) {
	do.Provide[embeddings.Embedder](i, func(i do.Injector) (embeddings.Embedder, error) {
		cfg := do.MustInvoke[*Config](i)
		logger := do.MustInvoke[logging.Logger](i)
		tracer := do.MustInvoke[tracing.Tracer](i)
		return ProvideEmbedder(context.Background(), cfg, logger, tracer)
	})
}
