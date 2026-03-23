package profiling

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNoopProvider(T *testing.T) {
	T.Parallel()

	T.Run("returns non-nil provider", func(t *testing.T) {
		t.Parallel()
		p := NewNoopProvider()
		assert.NotNil(t, p)
	})

	T.Run("Start returns nil", func(t *testing.T) {
		t.Parallel()
		p := NewNoopProvider()
		assert.NoError(t, p.Start(context.Background()))
	})

	T.Run("Shutdown returns nil", func(t *testing.T) {
		t.Parallel()
		p := NewNoopProvider()
		assert.NoError(t, p.Shutdown(context.Background()))
	})
}
