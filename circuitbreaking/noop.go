package circuitbreaking

// NoopCircuitBreaker is a no-op implementation that always allows operations to proceed.
type NoopCircuitBreaker struct{}

func (n *NoopCircuitBreaker) Failed() {}

func (n *NoopCircuitBreaker) Succeeded() {}

func (n *NoopCircuitBreaker) CanProceed() bool {
	return true
}

func (n *NoopCircuitBreaker) CannotProceed() bool {
	return false
}

// NewNoopCircuitBreaker returns a CircuitBreaker that always allows operations to proceed.
func NewNoopCircuitBreaker() CircuitBreaker {
	return &NoopCircuitBreaker{}
}
