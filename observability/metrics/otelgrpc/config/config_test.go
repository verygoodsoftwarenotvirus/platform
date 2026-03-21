package config

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
			CollectorEndpoint: t.Name(),
			ServiceName:       t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
