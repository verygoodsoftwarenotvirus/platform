package ratelimiting

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/errors"
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
