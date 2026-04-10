package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gocloud.dev/blob/memblob"
)

func TestProvideUploadManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
		}

		result := ProvideUploadManager(u)
		assert.NotNil(t, result)
		assert.Equal(t, u, result)
	})
}
