package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTracer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewTracerForTest(t.Name()))
	})
}

func TestNewNamedTracer(T *testing.T) {
	T.Parallel()

	T.Run("with nil provider", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNamedTracer(nil, t.Name()))
	})

	T.Run("with valid provider", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewNamedTracer(NewNoopTracerProvider(), t.Name()))
	})
}

func Test_otelSpanManager_StartSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		NewTracerForTest(t.Name()).StartSpan(t.Context())
	})
}

func Test_otelSpanManager_StartCustomSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		NewTracerForTest(t.Name()).StartCustomSpan(ctx, t.Name())
	})
}
