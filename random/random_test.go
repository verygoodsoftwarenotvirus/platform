package random

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type erroneousReader struct{}

func (r *erroneousReader) Read(p []byte) (n int, err error) {
	return -1, errors.New("blah")
}

func TestGenerateBase32EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		actual, err := GenerateBase32EncodedString(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestGenerateBase64EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		actual, err := GenerateBase64EncodedString(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestGenerateRawBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		actual, err := GenerateRawBytes(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
	})
}

func TestStandardSecretGenerator_GenerateBase32EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s := NewGenerator(nil, tracing.NewNoopTracerProvider())
		value, err := s.GenerateBase32EncodedString(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Greater(t, len(value), exampleLength)
		assert.NoError(t, err)
	})

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s, ok := NewGenerator(nil, tracing.NewNoopTracerProvider()).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateBase32EncodedString(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
	})
}

func TestStandardSecretGenerator_GenerateBase64EncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s := NewGenerator(nil, tracing.NewNoopTracerProvider())
		value, err := s.GenerateBase64EncodedString(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Greater(t, len(value), exampleLength)
		assert.NoError(t, err)
	})

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s, ok := NewGenerator(nil, tracing.NewNoopTracerProvider()).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateBase64EncodedString(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
	})
}

func TestStandardSecretGenerator_GenerateRawBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s := NewGenerator(nil, tracing.NewNoopTracerProvider())
		value, err := s.GenerateRawBytes(ctx, exampleLength)

		assert.NotEmpty(t, value)
		assert.Equal(t, len(value), exampleLength)
		assert.NoError(t, err)
	})

	T.Run("with error reading from secure PRNG", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleLength := 123

		s, ok := NewGenerator(nil, tracing.NewNoopTracerProvider()).(*standardGenerator)
		require.True(t, ok)
		s.randReader = &erroneousReader{}
		value, err := s.GenerateRawBytes(ctx, exampleLength)

		assert.Empty(t, value)
		assert.Error(t, err)
	})
}

func TestMustGenerateRawBytes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		result := MustGenerateRawBytes(ctx, 32)
		assert.NotEmpty(t, result)
	})
}

func TestGenerateHexEncodedString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		result, err := GenerateHexEncodedString(ctx, 32)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}
