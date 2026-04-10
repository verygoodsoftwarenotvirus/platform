package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackblazeB2Config_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &BackblazeB2Config{
			ApplicationKeyID: t.Name(),
			ApplicationKey:   t.Name(),
			BucketName:       t.Name(),
			Region:           t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing application key ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &BackblazeB2Config{
			ApplicationKey: t.Name(),
			BucketName:     t.Name(),
			Region:         t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing application key", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &BackblazeB2Config{
			ApplicationKeyID: t.Name(),
			BucketName:       t.Name(),
			Region:           t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing bucket name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &BackblazeB2Config{
			ApplicationKeyID: t.Name(),
			ApplicationKey:   t.Name(),
			Region:           t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing region", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &BackblazeB2Config{
			ApplicationKeyID: t.Name(),
			ApplicationKey:   t.Name(),
			BucketName:       t.Name(),
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
