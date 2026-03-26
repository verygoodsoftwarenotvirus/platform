package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopEmbedder_GenerateEmbedding(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		embedder := NewNoopEmbedder()

		result, err := embedder.GenerateEmbedding(ctx, &Input{
			Content: "hello world",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello world", result.SourceText)
		assert.Equal(t, "noop", result.Model)
		assert.Equal(t, "noop", result.Provider)
		assert.Equal(t, 0, result.Dimensions)
		assert.Empty(t, result.Vector)
		assert.False(t, result.GeneratedAt.IsZero())
	})
}
