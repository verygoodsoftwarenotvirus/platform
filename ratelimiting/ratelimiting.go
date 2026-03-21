package ratelimiting

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter limits the rate of operations per key.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Close() error
}

type inMemoryRateLimiter struct {
	limiters       sync.Map
	requestsPerSec float64
	burstSize      int
}

// NewInMemoryRateLimiter returns a RateLimiter that uses per-key limiters in memory.
func NewInMemoryRateLimiter(requestsPerSec float64, burstSize int) RateLimiter {
	return &inMemoryRateLimiter{
		requestsPerSec: requestsPerSec,
		burstSize:      burstSize,
	}
}

func (r *inMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	limiter := r.getOrCreateLimiter(ctx, key)
	return limiter.Allow(), nil
}

func (r *inMemoryRateLimiter) getOrCreateLimiter(_ context.Context, key string) *rate.Limiter {
	if v, ok := r.limiters.Load(key); ok {
		if x, ok2 := v.(*rate.Limiter); ok2 {
			return x
		}
	}

	limiter := rate.NewLimiter(rate.Limit(r.requestsPerSec), r.burstSize)
	if v, loaded := r.limiters.LoadOrStore(key, limiter); loaded {
		if x, ok2 := v.(*rate.Limiter); ok2 {
			return x
		}
	}

	return limiter
}

func (r *inMemoryRateLimiter) Close() error {
	return nil
}
