package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildArbitraryImage(T *testing.T) {
	T.Parallel()

	T.Run("returns image with correct dimensions", func(t *testing.T) {
		t.Parallel()
		img := BuildArbitraryImage(10)
		require.NotNil(t, img)
		bounds := img.Bounds()
		assert.Equal(t, 10, bounds.Dx())
		assert.Equal(t, 10, bounds.Dy())
	})

	T.Run("handles size 1", func(t *testing.T) {
		t.Parallel()
		img := BuildArbitraryImage(1)
		require.NotNil(t, img)
		assert.Equal(t, 1, img.Bounds().Dx())
	})
}

func TestBuildArbitraryImagePNGBytes(T *testing.T) {
	T.Parallel()

	T.Run("returns valid PNG bytes", func(t *testing.T) {
		t.Parallel()
		img, data := BuildArbitraryImagePNGBytes(5)
		require.NotNil(t, img)
		assert.NotEmpty(t, data)
		// PNG magic bytes
		assert.Equal(t, byte(0x89), data[0])
		assert.Equal(t, byte('P'), data[1])
		assert.Equal(t, byte('N'), data[2])
		assert.Equal(t, byte('G'), data[3])
	})
}

func TestBuildTestRequest(T *testing.T) {
	T.Parallel()

	T.Run("returns valid request", func(t *testing.T) {
		t.Parallel()
		req := BuildTestRequest(t)
		require.NotNil(t, req)
		assert.NotNil(t, req.Context())
	})
}
