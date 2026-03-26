package ratelimiting

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"

	"github.com/google/wire"
)

var (
	// Providers provides rate limiter construction for dependency injection.
	Providers = wire.NewSet(
		ProvideRateLimiterFromConfig,
	)
)

// ProvideRateLimiterFromConfig provides a RateLimiter from config.
func ProvideRateLimiterFromConfig(cfg *Config) (RateLimiter, error) {
	if cfg == nil {
		return NewNoopRateLimiter(), nil
	}
	limiter, err := cfg.ProvideRateLimiter()
	if err != nil {
		return nil, errors.Wrap(err, "provide rate limiter")
	}
	return limiter, nil
}
