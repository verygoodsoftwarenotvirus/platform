package stripe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripeConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			APIKey: "blah",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing API key", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			APIKey: "",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
