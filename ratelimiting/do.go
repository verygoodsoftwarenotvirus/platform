package ratelimiting

import (
	"github.com/samber/do/v2"
)

// RegisterRateLimiter registers a RateLimiter with the injector.
func RegisterRateLimiter(i do.Injector) {
	do.Provide[RateLimiter](i, func(i do.Injector) (RateLimiter, error) {
		return ProvideRateLimiterFromConfig(do.MustInvoke[*Config](i))
	})
}
