package ably

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			APIKey: "test.key:secret",
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing api key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}
