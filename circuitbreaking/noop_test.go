package circuitbreaking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNoopCircuitBreaker(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		assert.NotNil(t, x)
	})
}

func TestNoopCircuitBreaker_Failed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		x.Failed()
	})
}

func TestNoopCircuitBreaker_Succeeded(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		x.Succeeded()
	})
}

func TestNoopCircuitBreaker_CanProceed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		assert.True(t, x.CanProceed())
	})
}

func TestNoopCircuitBreaker_CannotProceed(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		assert.False(t, x.CannotProceed())
	})
}
