package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystemConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &FilesystemConfig{
			RootDirectory: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
