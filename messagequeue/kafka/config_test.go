package kafka

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
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty brokers", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			Brokers: []string{},
			GroupID: "test-group",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with nil brokers", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			GroupID: "test-group",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
