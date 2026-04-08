package embeddingscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterEmbedder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewTracerForTest("test"))
		do.ProvideValue(i, &Config{})

		RegisterEmbedder(i)

		embedder, err := do.Invoke[embeddings.Embedder](i)
		require.NoError(t, err)
		assert.NotNil(t, embedder)
	})
}
