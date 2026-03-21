package retry

import "context"

// NoopPolicy executes the operation exactly once with no retries.
type NoopPolicy struct{}

// Execute runs the operation once.
func (n *NoopPolicy) Execute(ctx context.Context, operation func(ctx context.Context) error) error {
	return operation(ctx)
}

// NewNoopPolicy returns a Policy that never retries.
func NewNoopPolicy() Policy {
	return &NoopPolicy{}
}
