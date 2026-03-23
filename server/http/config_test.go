package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			StartupDeadline: time.Second,
			Port:            8080,
			Debug:           true,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("returns error with missing port", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			StartupDeadline: time.Second,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("returns error with missing startup deadline", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Port: 8080,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("returns error with empty config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
