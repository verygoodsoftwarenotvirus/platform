package compression

import (
	"encoding/base64"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type whatever struct {
	Name string `json:"name"`
}

func TestNewCompressor(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		comp, err := NewCompressor(algoZstd)
		require.NoError(t, err)
		require.NotNil(t, comp)
	})

	T.Run("invalid algo", func(t *testing.T) {
		t.Parallel()

		comp, err := NewCompressor(algo(t.Name()))
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func Test_compressor_CompressBytes(T *testing.T) {
	T.Parallel()

	T.Run("zstandard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		comp, err := NewCompressor(algoZstd)
		require.NoError(t, err)

		x := &whatever{
			Name: "testing",
		}

		encoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

		expected := "KLUv_QQAmQAAeyJuYW1lIjoidGVzdGluZyJ9Ch6HXww="
		compressed, err := comp.CompressBytes(encoder.MustEncodeJSON(ctx, x))
		assert.NoError(t, err)
		actual := base64.URLEncoding.EncodeToString(compressed)

		assert.Equal(t, expected, actual)
	})

	T.Run("s2", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		comp, err := NewCompressor(algoS2)
		require.NoError(t, err)

		x := &whatever{
			Name: "testing",
		}

		encoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

		expected := "_wYAAFMyc1R3TwEXAABui7jXeyJuYW1lIjoidGVzdGluZyJ9Cg=="
		compressed, err := comp.CompressBytes(encoder.MustEncodeJSON(ctx, x))
		assert.NoError(t, err)
		actual := base64.URLEncoding.EncodeToString(compressed)

		assert.Equal(t, expected, actual)
	})

	T.Run("invalid algo", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		comp, err := NewCompressor(algoS2)
		require.NoError(t, err)

		comp.(*compressor).algo = "invalid"

		x := &whatever{
			Name: "testing",
		}

		encoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

		compressed, err := comp.CompressBytes(encoder.MustEncodeJSON(ctx, x))
		assert.Error(t, err)
		assert.Nil(t, compressed)
	})
}

func Test_compressor_DecompressBytes(T *testing.T) {
	T.Parallel()

	algorithms := []algo{
		algoZstd,
		algoS2,
	}

	for _, a := range algorithms {
		T.Run(string(a), func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			comp, err := NewCompressor(a)
			require.NoError(t, err)

			x := &whatever{
				Name: "testing",
			}

			encoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

			compressed, err := comp.CompressBytes(encoder.MustEncodeJSON(ctx, x))
			assert.NoError(t, err)

			decompressed, err := comp.DecompressBytes(compressed)
			assert.NoError(t, err)

			var y *whatever
			require.NoError(t, encoder.DecodeBytes(ctx, decompressed, &y))

			assert.Equal(t, x, y)
		})
	}

	T.Run("with invalid algo", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		comp, err := NewCompressor(algoZstd)
		require.NoError(t, err)

		x := &whatever{
			Name: "testing",
		}

		encoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

		compressed, err := comp.CompressBytes(encoder.MustEncodeJSON(ctx, x))
		assert.NoError(t, err)

		comp.(*compressor).algo = "invalid"

		decompressed, err := comp.DecompressBytes(compressed)
		assert.Error(t, err)
		assert.Nil(t, decompressed)
	})
}
