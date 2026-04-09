package qrcodes

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/boombuler/barcode"
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

	T.Run("with content exceeding QR capacity", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		b := NewBuilder("test-issuer", nil, nil)

		// A username longer than the maximum QR code capacity forces qr.Encode to fail.
		actual, err := b.BuildQRCode(ctx, strings.Repeat("a", 4000), "two-factor-secret")
		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	T.Run("with scale error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		b := NewBuilder("test-issuer", nil, nil).(*builder)
		b.scale = func(barcode.Barcode, int, int) (barcode.Barcode, error) {
			return nil, fmt.Errorf("scale error")
		}

		actual, err := b.BuildQRCode(ctx, "username", "two-factor-secret")
		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	T.Run("with png encode error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		b := NewBuilder("test-issuer", nil, nil).(*builder)
		b.pngEncode = func(*bytes.Buffer, barcode.Barcode) error {
			return fmt.Errorf("png encode error")
		}

		actual, err := b.BuildQRCode(ctx, "username", "two-factor-secret")
		assert.Empty(t, actual)
		assert.Error(t, err)
	})
}
