package objectstorage

import (
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

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

	T.Run("with missing bucket name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: MemoryProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   "invalid_provider",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with s3 provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   S3Provider,
			S3Config:   &S3Config{BucketName: t.Name()},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with s3 provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   S3Provider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with gcp provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   GCPCloudStorageProvider,
			GCP:        &GCPConfig{BucketName: t.Name()},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with gcp provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   GCPCloudStorageProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with r2 provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   R2Provider,
			R2Config: &R2Config{
				AccountID:       t.Name(),
				BucketName:      t.Name(),
				AccessKeyID:     t.Name(),
				SecretAccessKey: t.Name(),
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with r2 provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   R2Provider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with memory provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   MemoryProvider,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with filesystem provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   FilesystemProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with non-s3 provider having s3 config is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			BucketName: t.Name(),
			Provider:   MemoryProvider,
			S3Config:   &S3Config{BucketName: t.Name()},
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
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

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), cfg)
		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), nil)
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()
		cfg := &Config{}

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), cfg)
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("with filesystem provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()
		tempDir := os.TempDir()

		cfg := &Config{
			BucketName:       t.Name(),
			Provider:         FilesystemProvider,
			FilesystemConfig: &FilesystemConfig{RootDirectory: tempDir},
		}

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), cfg)
		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with bucket prefix", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		l := logging.NewNoopLogger()
		cfg := &Config{
			BucketName:   t.Name(),
			Provider:     MemoryProvider,
			BucketPrefix: "prefix/",
		}

		x, err := NewUploadManager(ctx, l, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), cfg)
		assert.NotNil(t, x)
		assert.NoError(t, err)
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

	T.Run("r2 happy path", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: R2Provider,
			R2Config: &R2Config{
				AccountID:       t.Name(),
				BucketName:      t.Name(),
				AccessKeyID:     t.Name(),
				SecretAccessKey: t.Name(),
			},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})

	T.Run("r2 with nil config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider: R2Provider,
			R2Config: nil,
		}

		assert.Error(t, u.selectBucket(ctx, cfg))
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

	T.Run("memory provider with bucket prefix", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		cfg := &Config{
			Provider:     MemoryProvider,
			BucketPrefix: "my-prefix/",
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
		assert.NotNil(t, u.bucket)
	})

	T.Run("unknown provider falls through to filesystem default", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := &Uploader{}
		tempDir := os.TempDir()
		cfg := &Config{
			Provider:         "something_unknown",
			FilesystemConfig: &FilesystemConfig{RootDirectory: tempDir},
		}

		assert.NoError(t, u.selectBucket(ctx, cfg))
	})
}
