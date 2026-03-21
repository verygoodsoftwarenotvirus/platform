package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type example struct {
	Name string `json:"name"`
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderMemory,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideCache(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideCache[example](&Config{
			Provider: ProviderMemory,
		})

		assert.NoError(t, err)
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideCache[example](&Config{})

		assert.Error(t, err)
	})
}
