package circuitbreaking

type NoopCircuitBreaker struct{}

func (n *NoopCircuitBreaker) Failed() {}

func (n *NoopCircuitBreaker) Succeeded() {}

func (n *NoopCircuitBreaker) CanProceed() bool {
	return true
}

func (n *NoopCircuitBreaker) CannotProceed() bool {
	return false
}

func NewNoopCircuitBreaker() CircuitBreaker {
	return &NoopCircuitBreaker{}
}
