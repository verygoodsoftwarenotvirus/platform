package noop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewCircuitBreaker()
		assert.NotNil(t, x)
	})
}

func TestCircuitBreaker_Failed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewCircuitBreaker()
		x.Failed()
	})
}

func TestCircuitBreaker_Succeeded(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewCircuitBreaker()
		x.Succeeded()
	})
}

func TestCircuitBreaker_CanProceed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewCircuitBreaker()
		assert.True(t, x.CanProceed())
	})
}

func TestCircuitBreaker_CannotProceed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewCircuitBreaker()
		assert.False(t, x.CannotProceed())
	})
}
