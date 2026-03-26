package config

import (
	"context"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/cache"
	"github.com/verygoodsoftwarenotvirus/platform/v4/cache/memory"
	"github.com/verygoodsoftwarenotvirus/platform/v4/cache/redis"
	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderMemory is the memory provider.
	ProviderMemory = "memory"
	// ProviderRedis is the redis provider.
	ProviderRedis = "redis"
)

type (
	// Config is the configuration for the cache.
	Config struct {
		Redis    *redis.Config `env:"init"     envPrefix:"REDIS_" json:"redis"`
		Provider string        `env:"PROVIDER" json:"provider"`
		Expiry   time.Duration `env:"EXPIRY"   envDefault:"1h"    json:"expiry"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ProviderMemory, ProviderRedis)),
		validation.Field(&cfg.Redis, validation.When(cfg.Provider == ProviderRedis, validation.Required)),
	)
}

// ProvideCache provides a Cache.
func ProvideCache[T any](cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, metricsProvider metrics.Provider) (cache.Cache[T], error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case ProviderMemory:
		return memory.NewInMemoryCache[T](logger, tracerProvider, metricsProvider)
	case ProviderRedis:
		return redis.NewRedisCache[T](cfg.Redis, time.Hour, logger, tracerProvider, metricsProvider)
	default:
		return nil, errors.Newf("invalid cache provider: %q", cfg.Provider)
	}
}
