package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/cache"
	mockcircuitbreaking "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/testutils"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

type mockRedisClient struct {
	mock.Mock
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.Called(ctx, key).Get(0).(*redis.StringCmd)
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return m.Called(ctx, key, value, expiration).Get(0).(*redis.StatusCmd)
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return m.Called(ctx, keys).Get(0).(*redis.IntCmd)
}

func (m *mockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	return m.Called(ctx).Get(0).(*redis.StatusCmd)
}

func gobEncodeExample(t *testing.T, e *example) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, gob.NewEncoder(&buf).Encode(e))

	return buf.String()
}

func buildTestImpl(t *testing.T) (*redisCacheImpl[example], *mockRedisClient, *mockcircuitbreaking.MockCircuitBreaker) {
	t.Helper()

	mp := metrics.NewNoopMetricsProvider()

	hitCounter, err := mp.NewInt64Counter("test_hits")
	require.NoError(t, err)

	missCounter, err := mp.NewInt64Counter("test_misses")
	require.NoError(t, err)

	setCounter, err := mp.NewInt64Counter("test_sets")
	require.NoError(t, err)

	delCounter, err := mp.NewInt64Counter("test_deletes")
	require.NoError(t, err)

	errCounter, err := mp.NewInt64Counter("test_errors")
	require.NoError(t, err)

	latencyHist, err := mp.NewFloat64Histogram("test_latency")
	require.NoError(t, err)

	client := &mockRedisClient{}
	cb := &mockcircuitbreaking.MockCircuitBreaker{}

	return &redisCacheImpl[example]{
		logger:           logging.NewNoopLogger(),
		tracer:           tracing.NewNamedTracer(nil, "test"),
		cacheHitCounter:  hitCounter,
		cacheMissCounter: missCounter,
		cacheSetCounter:  setCounter,
		cacheDelCounter:  delCounter,
		cacheErrCounter:  errCounter,
		latencyHist:      latencyHist,
		client:           client,
		circuitBreaker:   cb,
		expiration:       time.Minute,
	}, client, cb
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			QueueAddresses: []string{"localhost:6379"},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty addresses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			QueueAddresses: []string{},
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with nil addresses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestNewRedisCache(T *testing.T) {
	T.Parallel()

	T.Run("with single address", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	T.Run("with multiple addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379", "localhost:6380"}}

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	T.Run("with error creating cache hit counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating cache miss counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_misses", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating cache set counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_misses", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_sets", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating cache delete counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_misses", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_sets", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_deletes", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating cache error counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_misses", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_sets", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_deletes", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		noopMP := metrics.NewNoopMetricsProvider()
		h, histErr := noopMP.NewFloat64Histogram("test")
		require.NoError(t, histErr)

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_cache_hits", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_misses", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_sets", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_deletes", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_cache_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewFloat64Histogram", name+"_cache_latency_ms", []metric.Float64HistogramOption(nil)).Return(h, errors.New("histogram error"))

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		assert.Error(t, err)
		assert.Nil(t, c)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func Test_redisCacheImpl_Get_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		expected := &example{Name: t.Name()}
		encoded := gobEncodeExample(t, expected)

		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		cmd := redis.NewStringCmd(ctx)
		cmd.SetVal(encoded)
		client.On("Get", testutils.ContextMatcher, exampleKey).Return(cmd)

		actual, err := impl.Get(ctx, exampleKey)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, client, cb)
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(true)

		actual, err := impl.Get(ctx, exampleKey)
		assert.ErrorIs(t, err, cache.ErrNotFound)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(errors.New("redis error"))
		client.On("Get", testutils.ContextMatcher, exampleKey).Return(cmd)

		actual, err := impl.Get(ctx, exampleKey)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, client, cb)
	})

	T.Run("with decode error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)

		cmd := redis.NewStringCmd(ctx)
		cmd.SetVal("not valid gob data")
		client.On("Get", testutils.ContextMatcher, exampleKey).Return(cmd)

		actual, err := impl.Get(ctx, exampleKey)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, client, cb)
	})
}

func Test_redisCacheImpl_Set_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		cmd := redis.NewStatusCmd(ctx)
		cmd.SetVal("OK")
		client.On("Set", testutils.ContextMatcher, exampleKey, mock.AnythingOfType("string"), time.Minute).Return(cmd)

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, client, cb)
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(true)

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		cmd := redis.NewStatusCmd(ctx)
		cmd.SetErr(errors.New("redis error"))
		client.On("Set", testutils.ContextMatcher, exampleKey, mock.AnythingOfType("string"), time.Minute).Return(cmd)

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, client, cb)
	})
}

func Test_redisCacheImpl_Delete_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)
		cb.On("Succeeded").Return()

		cmd := redis.NewIntCmd(ctx)
		cmd.SetVal(1)
		client.On("Del", testutils.ContextMatcher, []string{exampleKey}).Return(cmd)

		err := impl.Delete(ctx, exampleKey)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, client, cb)
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(true)

		err := impl.Delete(ctx, exampleKey)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, cb)
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.On("CannotProceed").Return(false)
		cb.On("Failed").Return()

		cmd := redis.NewIntCmd(ctx)
		cmd.SetErr(errors.New("redis error"))
		client.On("Del", testutils.ContextMatcher, []string{exampleKey}).Return(cmd)

		err := impl.Delete(ctx, exampleKey)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, client, cb)
	})
}

func Test_redisCacheImpl_Ping_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, _ := buildTestImpl(t)

		cmd := redis.NewStatusCmd(ctx)
		cmd.SetVal("PONG")
		client.On("Ping", testutils.ContextMatcher).Return(cmd)

		assert.NoError(t, impl.Ping(ctx))

		mock.AssertExpectationsForObjects(t, client)
	})

	T.Run("with error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, _ := buildTestImpl(t)

		cmd := redis.NewStatusCmd(ctx)
		cmd.SetErr(errors.New("connection refused"))
		client.On("Ping", testutils.ContextMatcher).Return(cmd)

		assert.Error(t, impl.Ping(ctx))

		mock.AssertExpectationsForObjects(t, client)
	})
}

func Test_buildRedisClient(T *testing.T) {
	T.Parallel()

	T.Run("with single address", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			QueueAddresses: []string{"localhost:6379"},
			Username:       "user",
			Password:       "pass",
		}

		c := buildRedisClient(cfg)
		assert.NotNil(t, c)
	})

	T.Run("with multiple addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			QueueAddresses: []string{"localhost:6379", "localhost:6380"},
			Username:       "user",
			Password:       "pass",
		}

		c := buildRedisClient(cfg)
		assert.NotNil(t, c)
	})

	T.Run("with no addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			QueueAddresses: []string{},
		}

		c := buildRedisClient(cfg)
		assert.Nil(t, c)
	})
}
