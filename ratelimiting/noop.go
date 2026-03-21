package ratelimiting

import "context"

// NoopRateLimiter always allows requests.
type NoopRateLimiter struct{}

// Allow always returns true.
func (n *NoopRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return true, nil
}

// Close is a no-op.
func (n *NoopRateLimiter) Close() error {
	return nil
}

// NewNoopRateLimiter returns a RateLimiter that never limits.
func NewNoopRateLimiter() RateLimiter {
	return &NoopRateLimiter{}
}
