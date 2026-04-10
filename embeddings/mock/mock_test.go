package mock

import (
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/embeddings"

	"github.com/stretchr/testify/require"
)

func TestEmbedder_GenerateEmbedding(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &Embedder{}
		input := &embeddings.Input{Content: "hello", Model: "test"}
		expected := &embeddings.Embedding{
			SourceText: "hello",
			Model:      "test",
			Provider:   "mock",
		}

		m.On("GenerateEmbedding", t.Context(), input).Return(expected, nil)

		ctx := t.Context()
		result, err := m.GenerateEmbedding(ctx, input)

		require.NoError(t, err)
		require.Equal(t, expected, result)
		m.AssertExpectations(t)
	})

	T.Run("with nil result", func(t *testing.T) {
		t.Parallel()

		m := &Embedder{}
		input := &embeddings.Input{Content: "hello", Model: "test"}

		m.On("GenerateEmbedding", t.Context(), input).Return(nil, fmt.Errorf("embedding failed"))

		ctx := t.Context()
		result, err := m.GenerateEmbedding(ctx, input)

		require.Error(t, err)
		require.Nil(t, result)
		m.AssertExpectations(t)
	})
}
