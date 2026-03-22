package httpclient

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

		assert.Equal(t, defaultTimeout, cfg.Timeout)
		assert.Equal(t, defaultMaxIdleConns, cfg.MaxIdleConns)
		assert.Equal(t, defaultMaxIdleConnsPerHost, cfg.MaxIdleConnsPerHost)
	})

	T.Run("preserves non-zero values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Timeout:             5 * time.Second,
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 25,
		}
		cfg.EnsureDefaults()

		assert.Equal(t, 5*time.Second, cfg.Timeout)
		assert.Equal(t, 50, cfg.MaxIdleConns)
		assert.Equal(t, 25, cfg.MaxIdleConnsPerHost)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Timeout:             time.Second,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
		}

		err := cfg.ValidateWithContext(ctx)
		require.NoError(t, err)
	})

	T.Run("invalid timeout", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Timeout:             0,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})

	T.Run("invalid MaxIdleConns", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Timeout:             time.Second,
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: 5,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})

	T.Run("invalid MaxIdleConnsPerHost", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cfg := &Config{
			Timeout:             time.Second,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 0,
		}

		err := cfg.ValidateWithContext(ctx)
		require.Error(t, err)
	})
}
