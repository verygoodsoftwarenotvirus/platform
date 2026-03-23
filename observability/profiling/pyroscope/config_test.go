package pyroscope

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	ctx := context.Background()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			ServerAddress: "http://localhost:4040",
			UploadRate:    15 * time.Second,
		}
		assert.NoError(t, c.ValidateWithContext(ctx))
	})

	T.Run("missing server address", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			UploadRate: 15 * time.Second,
		}
		assert.Error(t, c.ValidateWithContext(ctx))
	})

	T.Run("missing upload rate", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			ServerAddress: "http://localhost:4040",
		}
		assert.Error(t, c.ValidateWithContext(ctx))
	})
}
