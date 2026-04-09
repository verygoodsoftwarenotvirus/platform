package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/uploads/objectstorage"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Storage: objectstorage.Config{
				FilesystemConfig:  &objectstorage.FilesystemConfig{RootDirectory: "/blah"},
				S3Config:          &objectstorage.S3Config{BucketName: "blahs"},
				BucketName:        "blahs",
				UploadFilenameKey: "blahs",
				Provider:          "blahs",
			},
			Debug: false,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty storage", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
