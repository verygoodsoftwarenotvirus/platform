package objectstorage

import (
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName:       t.Name(),
			Provider:         FilesystemProvider,
			FilesystemConfig: &FilesystemConfig{RootDirectory: t.Name()},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestNewUploadManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   MemoryProvider,
		}

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), cfg)
		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), nil)
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()
		cfg := &Config{}

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), cfg)
		assert.Nil(t, x)
		assert.Error(t, err)
	})
}

func TestUploader_selectBucket(T *testing.T) {
	T.Parallel()

	T.Run("s3 happy path", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: S3Provider,
			S3Config: &S3Config{
				BucketName: t.Name(),
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("s3 with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: S3Provider,
			S3Config: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})

	T.Run("memory provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: MemoryProvider,
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("filesystem happy path", func(t *testing.T) {
		t.Parallel()

		tempDir := os.TempDir()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: FilesystemProvider,
			FilesystemConfig: &FilesystemConfig{
				RootDirectory: tempDir,
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("filesystem with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider:         FilesystemProvider,
			FilesystemConfig: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
	})
}
