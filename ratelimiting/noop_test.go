package ratelimiting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	limiter := NewNoopRateLimiter()
	ctx := context.Background()

	for range 100 {
		allowed, err := limiter.Allow(ctx, "any")
		require.NoError(t, err)
		assert.True(t, allowed)
	}
}

func TestNoopRateLimiter_Close(t *testing.T) {
	t.Parallel()

	limiter := NewNoopRateLimiter()
	err := limiter.Close()
	require.NoError(t, err)
}
