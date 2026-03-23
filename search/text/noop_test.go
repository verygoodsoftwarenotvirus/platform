package textsearch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopIndexManager_Search(T *testing.T) {
	T.Parallel()

	T.Run("returns empty slice and no error", func(t *testing.T) {
		t.Parallel()

		m := &NoopIndexManager[string]{}
		results, err := m.Search(context.Background(), "query")

		require.NoError(t, err)
		assert.Empty(t, results)
		assert.NotNil(t, results)
	})
}

func TestNoopIndexManager_Index(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := &NoopIndexManager[string]{}
		err := m.Index(context.Background(), "id", "value")

		assert.NoError(t, err)
	})
}

func TestNoopIndexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := &NoopIndexManager[string]{}
		err := m.Delete(context.Background(), "id")

		assert.NoError(t, err)
	})
}

func TestNoopIndexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("returns no error", func(t *testing.T) {
		t.Parallel()

		m := &NoopIndexManager[string]{}
		err := m.Wipe(context.Background())

		assert.NoError(t, err)
	})
}
