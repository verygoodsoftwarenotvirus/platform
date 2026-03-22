package circuitbreaking

import "testing"

func TestNoopCircuitBreaker_Obligatory(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := NewNoopCircuitBreaker()
		x.Failed()
		x.Succeeded()
		x.CanProceed()
		x.CannotProceed()
	})
}
