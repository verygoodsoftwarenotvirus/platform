package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("sets defaults for zero values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()

		assert.Equal(t, uint(3), cfg.MaxAttempts)
		assert.Equal(t, 100*time.Millisecond, cfg.InitialDelay)
		assert.Equal(t, 5*time.Second, cfg.MaxDelay)
		assert.Equal(t, 2.0, cfg.Multiplier)
	})

	T.Run("preserves non-zero values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			MaxAttempts:  7,
			InitialDelay: 1 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   3.0,
		}
		cfg.EnsureDefaults()

		assert.Equal(t, uint(7), cfg.MaxAttempts)
		assert.Equal(t, 1*time.Second, cfg.InitialDelay)
		assert.Equal(t, 10*time.Second, cfg.MaxDelay)
		assert.Equal(t, 3.0, cfg.Multiplier)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			MaxAttempts:  1,
			InitialDelay: time.Millisecond,
			MaxDelay:     time.Second,
			Multiplier:   2.0,
		}

		err := cfg.ValidateWithContext(ctx)
		require.NoError(t, err)
	})

	T.Run("invalid MaxAttempts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			MaxAttempts:  0,
			InitialDelay: time.Millisecond,
			MaxDelay:     time.Second,
			Multiplier:   2.0,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})

	T.Run("invalid InitialDelay", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			MaxAttempts:  1,
			InitialDelay: 0,
			MaxDelay:     time.Second,
			Multiplier:   2.0,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})

	T.Run("invalid Multiplier", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			MaxAttempts:  1,
			InitialDelay: time.Millisecond,
			MaxDelay:     time.Second,
			Multiplier:   0.5,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})
}
