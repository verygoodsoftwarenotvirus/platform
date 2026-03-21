package ratelimiting

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_EnsureDefaults(t *testing.T) {
	t.Parallel()

	t.Run("sets defaults for zero values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()

		assert.Equal(t, 10.0, cfg.RequestsPerSec)
		assert.Equal(t, 20, cfg.BurstSize)
	})

	t.Run("preserves non-zero values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			RequestsPerSec: 5.0,
			BurstSize:      10,
		}
		cfg.EnsureDefaults()

		assert.Equal(t, 5.0, cfg.RequestsPerSec)
		assert.Equal(t, 10, cfg.BurstSize)
	})
}

func TestConfig_ProvideRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("nil config returns noop", func(t *testing.T) {
		t.Parallel()

		var cfg *Config
		limiter, err := cfg.ProvideRateLimiter()
		require.NoError(t, err)
		require.NotNil(t, limiter)

		allowed, err := limiter.Allow(context.Background(), "x")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("empty provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ""}
		limiter, err := cfg.ProvideRateLimiter()
		require.NoError(t, err)
		require.NotNil(t, limiter)

		allowed, err := limiter.Allow(context.Background(), "x")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("noop provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderNoop}
		limiter, err := cfg.ProvideRateLimiter()
		require.NoError(t, err)
		require.NotNil(t, limiter)

		allowed, err := limiter.Allow(context.Background(), "x")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("memory provider returns in-memory limiter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:       ProviderMemory,
			RequestsPerSec: 1,
			BurstSize:      1,
		}
		limiter, err := cfg.ProvideRateLimiter()
		require.NoError(t, err)
		require.NotNil(t, limiter)

		allowed, err := limiter.Allow(context.Background(), "x")
		require.NoError(t, err)
		assert.True(t, allowed)

		allowed, err = limiter.Allow(context.Background(), "x")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("unknown provider returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: "unknown"}
		limiter, err := cfg.ProvideRateLimiter()
		require.Error(t, err)
		assert.Nil(t, limiter)
		assert.Contains(t, err.Error(), "unknown")
	})
}

func TestConfig_ValidateWithContext(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			RequestsPerSec: 1.0,
			BurstSize:      1,
		}

		err := cfg.ValidateWithContext(ctx)
		require.NoError(t, err)
	})

	t.Run("invalid RequestsPerSec", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			RequestsPerSec: -1,
			BurstSize:      1,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})

	t.Run("invalid BurstSize", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			RequestsPerSec: 1.0,
			BurstSize:      -1,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})
}
