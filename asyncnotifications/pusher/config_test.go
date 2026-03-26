package pusher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			AppID:   "123",
			Key:     "key",
			Secret:  "secret",
			Cluster: "us2",
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing required fields", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}
