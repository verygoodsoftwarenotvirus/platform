package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			APIKey: "test-key",
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}
