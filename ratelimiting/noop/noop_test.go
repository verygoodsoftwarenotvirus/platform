package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_Allow(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		limiter := NewRateLimiter()
		ctx := context.Background()

		for range 100 {
			allowed, err := limiter.Allow(ctx, "any")
			require.NoError(t, err)
			assert.True(t, allowed)
		}
	})
}

func TestRateLimiter_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		limiter := NewRateLimiter()
		err := limiter.Close()
		require.NoError(t, err)
	})
}
