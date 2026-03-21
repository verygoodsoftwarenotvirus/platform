package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInstrumentedSQLTracer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewInstrumentedSQLTracer(NewNoopTracerProvider(), t.Name()))
	})
}

func Test_instrumentedSQLTracerWrapper_GetSpan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		w := NewInstrumentedSQLTracer(NewNoopTracerProvider(), t.Name())

		assert.NotNil(t, w.GetSpan(ctx))
	})
}
