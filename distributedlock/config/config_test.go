package distributedlockcfg

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock"
	pglock "github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock/postgres"
	redislock "github.com/verygoodsoftwarenotvirus/platform/v5/distributedlock/redis"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("redis provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Provider: RedisProvider,
			Redis: &redislock.Config{
				Addresses: []string{"localhost:6379"},
				KeyPrefix: "lock:",
			},
		}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("postgres provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Provider: PostgresProvider,
			Postgres: &pglock.Config{},
		}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("memory provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: MemoryProvider}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("noop provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: NoopProvider}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("redis without config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: RedisProvider}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("postgres without config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: PostgresProvider}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: "made-up"}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("empty provider is valid (noop)", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestProvideLocker(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideLocker(
			t.Context(),
			nil,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		assert.ErrorIs(t, err, distributedlock.ErrNilConfig)
	})

	T.Run("memory provider returns a working locker", func(t *testing.T) {
		t.Parallel()
		l, err := ProvideLocker(
			t.Context(),
			&Config{Provider: MemoryProvider},
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, l)
		lock, err := l.Acquire(t.Context(), "k", time.Second)
		require.NoError(t, err)
		require.NoError(t, lock.Release(t.Context()))
	})

	T.Run("noop provider", func(t *testing.T) {
		t.Parallel()
		l, err := ProvideLocker(
			t.Context(),
			&Config{Provider: NoopProvider},
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, l)
	})

	T.Run("unknown provider returns noop", func(t *testing.T) {
		t.Parallel()
		l, err := ProvideLocker(
			t.Context(),
			&Config{Provider: "unknown"},
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, l)
	})

	T.Run("empty provider returns noop", func(t *testing.T) {
		t.Parallel()
		l, err := ProvideLocker(
			t.Context(),
			&Config{},
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, l)
	})

	T.Run("provider with whitespace returns noop", func(t *testing.T) {
		t.Parallel()
		l, err := ProvideLocker(
			t.Context(),
			&Config{Provider: "   "},
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
		)
		require.NoError(t, err)
		require.NotNil(t, l)
	})
}
