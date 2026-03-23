package circuitbreaking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              0.99,
			MinimumSampleThreshold: 123,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})

	T.Run("with missing name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:      "",
			ErrorRate: 0.99,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.Error(t, err)
	})

	T.Run("with error rate exceeding max", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:      t.Name(),
			ErrorRate: 200,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.Error(t, err)
	})
}

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("with empty config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()

		assert.Equal(t, "UNKNOWN", cfg.Name)
		assert.Equal(t, float64(100), cfg.ErrorRate)
		assert.Equal(t, uint64(1_000_000), cfg.MinimumSampleThreshold)
	})

	T.Run("does not override set values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Name:                   "test",
			ErrorRate:              50.0,
			MinimumSampleThreshold: 500,
		}
		cfg.EnsureDefaults()

		assert.Equal(t, "test", cfg.Name)
		assert.Equal(t, 50.0, cfg.ErrorRate)
		assert.Equal(t, uint64(500), cfg.MinimumSampleThreshold)
	})
}
