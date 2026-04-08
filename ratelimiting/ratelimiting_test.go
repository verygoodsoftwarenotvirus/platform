package ratelimiting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRateLimiter_Allow(T *testing.T) {
	T.Parallel()

	T.Run("allows within burst", func(t *testing.T) {
		t.Parallel()

		limiter, err := NewInMemoryRateLimiter(nil, 10, 3)
		require.NoError(t, err)
		defer limiter.Close()

		ctx := context.Background()

		allowed, err := limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	T.Run("different keys have independent limits", func(t *testing.T) {
		t.Parallel()

		limiter, err := NewInMemoryRateLimiter(nil, 10, 1)
		require.NoError(t, err)
		defer limiter.Close()

		ctx := context.Background()

		allowed, err := limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, "key2")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(ctx, "key1")
		require.NoError(t, err)
		assert.False(t, allowed)

		allowed, err = limiter.Allow(ctx, "key2")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	T.Run("Close is safe", func(t *testing.T) {
		t.Parallel()

		limiter, err := NewInMemoryRateLimiter(nil, 10, 1)
		require.NoError(t, err)
		err = limiter.Close()
		require.NoError(t, err)
	})
}
