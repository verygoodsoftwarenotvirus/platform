package algolia

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(T *testing.T) {
	T.Parallel()

	T.Run("zero value", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		assert.Empty(t, cfg.AppID)
		assert.Empty(t, cfg.APIKey)
		assert.Equal(t, time.Duration(0), cfg.Timeout)
	})

	T.Run("with values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			AppID:   "test-app-id",
			APIKey:  "test-api-key",
			Timeout: 5 * time.Second,
		}

		assert.Equal(t, "test-app-id", cfg.AppID)
		assert.Equal(t, "test-api-key", cfg.APIKey)
		assert.Equal(t, 5*time.Second, cfg.Timeout)
	})
}
