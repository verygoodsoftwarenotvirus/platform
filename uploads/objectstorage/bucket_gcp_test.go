package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCPConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &GCPConfig{
			BucketName: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing bucket name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &GCPConfig{}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
