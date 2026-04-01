package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestR2Config_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &R2Config{
			AccountID:       t.Name(),
			BucketName:      t.Name(),
			AccessKeyID:     t.Name(),
			SecretAccessKey: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
