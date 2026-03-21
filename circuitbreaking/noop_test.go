package circuitbreaking

import "testing"

func TestNoopCircuitBreaker_Obligatory(t *testing.T) {
	t.Parallel()

	x := NewNoopCircuitBreaker()
	x.Failed()
	x.Succeeded()
	x.CanProceed()
	x.CannotProceed()
}
