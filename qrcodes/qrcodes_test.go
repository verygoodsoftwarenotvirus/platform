package qrcodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		b := NewBuilder("test-issuer", nil, nil)
		assert.NotNil(t, b)
	})
}

func Test_builder_BuildQRCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		b := NewBuilder("test-issuer", nil, nil)

		actual, err := b.BuildQRCode(ctx, "username", "two-factor-secret")
		require.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}
