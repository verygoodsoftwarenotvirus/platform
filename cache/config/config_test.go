package config

import (
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/cache/redis"
	circuitbreakingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

type example struct {
	Name string `json:"name"`
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("memory provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderMemory,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("redis provider with config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderRedis,
			Redis:    &redis.Config{QueueAddresses: []string{"localhost:6379"}},
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("redis provider missing config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderRedis}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("invalid provider name", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: "vault"}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestProvideCache(T *testing.T) {
	T.Parallel()

	T.Run("memory provider", func(t *testing.T) {
		t.Parallel()

		c, err := ProvideCache[example](t.Context(), &Config{
			Provider: ProviderMemory,
		}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())

		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	T.Run("redis provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderRedis,
			Redis:    &redis.Config{QueueAddresses: []string{"localhost:6379"}},
		}
		cfg.CircuitBreaker.Name = "cache-breaker"

		c, err := ProvideCache[example](
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
		)

		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	T.Run("redis provider with cluster addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderRedis,
			Redis:    &redis.Config{QueueAddresses: []string{"localhost:6379", "localhost:6380"}},
		}
		cfg.CircuitBreaker.Name = "cache-breaker-cluster"

		c, err := ProvideCache[example](
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
		)

		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	T.Run("redis provider with circuit breaker error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderRedis,
			Redis:    &redis.Config{QueueAddresses: []string{"localhost:6379"}},
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "redis-cache-breaker",
				ErrorRate:              50,
				MinimumSampleThreshold: 10,
			},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", "redis-cache-breaker_circuit_breaker_tripped", []metric.Int64CounterOption(nil)).
			Return(&mockmetrics.Int64Counter{}, fmt.Errorf("counter init failure"))

		c, err := ProvideCache[example](
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
		)

		require.Error(t, err)
		assert.Nil(t, c)
		mp.AssertExpectations(t)
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideCache[example](t.Context(), &Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())

		assert.Error(t, err)
	})
}
