package noop

import (
	"testing"

	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type example struct {
	Name string
}

func TestNewIndex(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex[example]()
		assert.NotNil(t, idx)
	})
}

func TestIndexManager_Upsert(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex[example]()
		require.NoError(t, idx.Upsert(t.Context(), vectorsearch.Vector[example]{
			ID:        "abc",
			Embedding: []float32{0.1, 0.2, 0.3},
			Metadata:  &example{Name: "doc"},
		}))
	})
}

func TestIndexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex[example]()
		require.NoError(t, idx.Delete(t.Context(), "abc", "def"))
	})
}

func TestIndexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex[example]()
		require.NoError(t, idx.Wipe(t.Context()))
	})
}

func TestIndexManager_Query(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		idx := NewIndex[example]()
		results, err := idx.Query(t.Context(), vectorsearch.QueryRequest{
			Embedding: []float32{0.1, 0.2, 0.3},
			TopK:      10,
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
