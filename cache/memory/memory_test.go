package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleKey = "example"
)

type example struct {
	Name string `json:"name"`
}

func Test_newInMemoryCache(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual, err := NewInMemoryCache[example](nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func Test_inMemoryCacheImpl_Get(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, err := NewInMemoryCache[example](nil, nil, nil)
		require.NoError(t, err)

		expected := &example{Name: t.Name()}
		assert.NoError(t, c.Set(ctx, exampleKey, expected))

		actual, err := c.Get(ctx, exampleKey)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})
}

func Test_inMemoryCacheImpl_Set(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, err := NewInMemoryCache[example](nil, nil, nil)
		require.NoError(t, err)

		assert.Len(t, c.(*inMemoryCacheImpl[example]).cache, 0)
		assert.NoError(t, c.Set(ctx, exampleKey, &example{Name: t.Name()}))
		assert.Len(t, c.(*inMemoryCacheImpl[example]).cache, 1)
	})
}

func Test_inMemoryCacheImpl_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c, err := NewInMemoryCache[example](nil, nil, nil)
		require.NoError(t, err)

		assert.Len(t, c.(*inMemoryCacheImpl[example]).cache, 0)
		assert.NoError(t, c.Set(ctx, exampleKey, &example{Name: t.Name()}))
		assert.Len(t, c.(*inMemoryCacheImpl[example]).cache, 1)
		assert.NoError(t, c.Delete(ctx, exampleKey))
		assert.Len(t, c.(*inMemoryCacheImpl[example]).cache, 0)
	})
}

func Test_inMemoryCacheImpl_Ping(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, err := NewInMemoryCache[example](nil, nil, nil)
		require.NoError(t, err)
		assert.NoError(t, c.Ping(t.Context()))
	})
}
