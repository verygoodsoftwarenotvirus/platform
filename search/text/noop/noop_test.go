package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexManager_Search(T *testing.T) {
	T.Parallel()

	T.Run("returns empty slice and no error", func(t *testing.T) {
		t.Parallel()

		m := NewIndexManager[string]()
		results, err := m.Search(context.Background(), "query")

		require.NoError(t, err)
		assert.Empty(t, results)
		assert.NotNil(t, results)
	})
}

func TestIndexManager_Index(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := NewIndexManager[string]()
		err := m.Index(context.Background(), "id", "value")

		assert.NoError(t, err)
	})
}

func TestIndexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := NewIndexManager[string]()
		err := m.Delete(context.Background(), "id")

		assert.NoError(t, err)
	})
}

func TestIndexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := NewIndexManager[string]()
		err := m.Wipe(context.Background())

		assert.NoError(t, err)
	})
}
