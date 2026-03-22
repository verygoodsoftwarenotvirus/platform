package objectstorage

import (
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob/memblob"
)

func TestUploader_ReadFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, []byte(t.Name()), nil))

		u := &Uploader{
			bucket: b,
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.NoError(t, err)
		assert.NotNil(t, x)
	})

	T.Run("with invalid file", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleFilename := "hello_world.txt"

		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.Error(t, err)
		assert.Nil(t, x)
	})
}

func TestUploader_SaveFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		ctx := t.Context()
		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		assert.NoError(t, u.SaveFile(ctx, tempFile.Name(), []byte(t.Name())))
	})
}
