package ratelimiting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("allows within burst", func(t *testing.T) {
		t.Parallel()

		limiter := NewInMemoryRateLimiter(10, 3)
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

	t.Run("different keys have independent limits", func(t *testing.T) {
		t.Parallel()

		limiter := NewInMemoryRateLimiter(10, 1)
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

	t.Run("Close is safe", func(t *testing.T) {
		t.Parallel()

		limiter := NewInMemoryRateLimiter(10, 1)
		err := limiter.Close()
		require.NoError(t, err)
	})
}
