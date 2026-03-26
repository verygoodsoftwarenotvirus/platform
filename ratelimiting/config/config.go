package ratelimitingcfg

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/ratelimiting"
	"github.com/verygoodsoftwarenotvirus/platform/v4/ratelimiting/noop"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ProviderMemory = "memory"
	ProviderNoop   = "noop"

	defaultRequestsPerSec = 10.0
	defaultBurstSize      = 20
)

// Config configures rate limiting.
type Config struct {
	Provider       string  `env:"PROVIDER"         json:"provider"`
	RequestsPerSec float64 `env:"REQUESTS_PER_SEC" json:"requestsPerSecond"`
	BurstSize      int     `env:"BURST_SIZE"       json:"burstSize"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// EnsureDefaults sets default values for zero fields.
func (cfg *Config) EnsureDefaults() {
	if cfg.RequestsPerSec == 0 {
		cfg.RequestsPerSec = defaultRequestsPerSec
	}
	if cfg.BurstSize == 0 {
		cfg.BurstSize = defaultBurstSize
	}
}

// ValidateWithContext validates the config.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.RequestsPerSec, validation.Min(0.0)),
		validation.Field(&cfg.BurstSize, validation.Min(0)),
	)
}

// ProvideRateLimiter returns a RateLimiter from config.
func (cfg *Config) ProvideRateLimiter() (ratelimiting.RateLimiter, error) {
	if cfg == nil {
		return noop.NewRateLimiter(), nil
	}
	cfg.EnsureDefaults()

	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case "", ProviderNoop:
		return noop.NewRateLimiter(), nil
	case ProviderMemory:
		return ratelimiting.NewInMemoryRateLimiter(cfg.RequestsPerSec, cfg.BurstSize), nil
	default:
		return nil, errors.Newf("unknown rate limiter provider: %q", cfg.Provider)
	}
}

// ProvideRateLimiterFromConfig provides a RateLimiter from config.
func ProvideRateLimiterFromConfig(cfg *Config) (ratelimiting.RateLimiter, error) {
	if cfg == nil {
		return noop.NewRateLimiter(), nil
	}
	limiter, err := cfg.ProvideRateLimiter()
	if err != nil {
		return nil, errors.Wrap(err, "provide rate limiter")
	}
	return limiter, nil
}
