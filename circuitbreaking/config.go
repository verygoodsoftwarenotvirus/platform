package circuitbreaking

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Name                   string  `env:"NAME"                     json:"name"`
	ErrorRate              float64 `env:"ERROR_RATE"               json:"circuitBreakerErrorPercentage"`
	MinimumSampleThreshold uint64  `env:"MINIMUM_SAMPLE_THRESHOLD" json:"circuitBreakerMinimumOccurrenceThreshold"`
}

func (cfg *Config) EnsureDefaults() {
	if cfg.Name == "" {
		cfg.Name = "UNKNOWN"
	}

	if cfg.ErrorRate == 0 {
		cfg.ErrorRate = 100
	}

	if cfg.MinimumSampleThreshold == 0 {
		cfg.MinimumSampleThreshold = 1_000_000
	}
}

func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Name, validation.Required),
		validation.Field(&cfg.ErrorRate, validation.Min(0.00), validation.Max(100.0)),
		validation.Field(&cfg.MinimumSampleThreshold),
	)
}
