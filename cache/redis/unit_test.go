package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/cache"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/go-redis/redis/v8"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"go.opentelemetry.io/otel/metric"
)

func gobEncodeExample(t *testing.T, e *example) string {
	t.Helper()

	var buf bytes.Buffer
	must.NoError(t, gob.NewEncoder(&buf).Encode(e))

	return buf.String()
}

func buildTestImpl(t *testing.T) (*redisCacheImpl[example], *redisClientMock, *CircuitBreakerMock) {
	t.Helper()

	mp := metrics.NewNoopMetricsProvider()

	hitCounter, err := mp.NewInt64Counter("test_hits")
	must.NoError(t, err)

	missCounter, err := mp.NewInt64Counter("test_misses")
	must.NoError(t, err)

	setCounter, err := mp.NewInt64Counter("test_sets")
	must.NoError(t, err)

	delCounter, err := mp.NewInt64Counter("test_deletes")
	must.NoError(t, err)

	errCounter, err := mp.NewInt64Counter("test_errors")
	must.NoError(t, err)

	latencyHist, err := mp.NewFloat64Histogram("test_latency")
	must.NoError(t, err)

	client := &redisClientMock{}
	cb := &CircuitBreakerMock{}

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

// counterResult bundles the values a mocked NewInt64Counter call returns.
type counterResult struct {
	counter metrics.Int64Counter
	err     error
}

// newCounterProviderMock returns a ProviderMock whose NewInt64Counter implementation
// looks up the result keyed on the counter name. Unknown names fail the test.
func newCounterProviderMock(t *testing.T, results map[string]counterResult) *ProviderMock {
	t.Helper()
	return &ProviderMock{
		NewInt64CounterFunc: func(metricName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
			res, ok := results[metricName]
			if !ok {
				t.Fatalf("unexpected NewInt64Counter call: %q", metricName)
			}
			return res.counter, res.err
		},
	}
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			QueueAddresses: []string{"localhost:6379"},
		}

		test.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty addresses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			QueueAddresses: []string{},
		}

		test.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with nil addresses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{}

		test.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestNewRedisCache(T *testing.T) {
	T.Parallel()

	okCounter := func() metrics.Int64Counter { return metrics.Int64CounterForTest(T, "x") }

	T.Run("with single address", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, nil, nil)
		must.NoError(t, err)
		test.NotNil(t, c)
	})

	T.Run("with multiple addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379", "localhost:6380"}}

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, nil, nil)
		must.NoError(t, err)
		test.NotNil(t, c)
	})

	T.Run("with error creating cache hit counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := newCounterProviderMock(t, map[string]counterResult{
			name + "_cache_hits": {counter: okCounter(), err: errors.New("counter error")},
		})

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating cache miss counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := newCounterProviderMock(t, map[string]counterResult{
			name + "_cache_hits":   {counter: okCounter()},
			name + "_cache_misses": {counter: okCounter(), err: errors.New("counter error")},
		})

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating cache set counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := newCounterProviderMock(t, map[string]counterResult{
			name + "_cache_hits":   {counter: okCounter()},
			name + "_cache_misses": {counter: okCounter()},
			name + "_cache_sets":   {counter: okCounter(), err: errors.New("counter error")},
		})

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 3, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating cache delete counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := newCounterProviderMock(t, map[string]counterResult{
			name + "_cache_hits":    {counter: okCounter()},
			name + "_cache_misses":  {counter: okCounter()},
			name + "_cache_sets":    {counter: okCounter()},
			name + "_cache_deletes": {counter: okCounter(), err: errors.New("counter error")},
		})

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 4, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating cache error counter", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		mp := newCounterProviderMock(t, map[string]counterResult{
			name + "_cache_hits":    {counter: okCounter()},
			name + "_cache_misses":  {counter: okCounter()},
			name + "_cache_sets":    {counter: okCounter()},
			name + "_cache_deletes": {counter: okCounter()},
			name + "_cache_errors":  {counter: okCounter(), err: errors.New("counter error")},
		})

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 5, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{QueueAddresses: []string{"localhost:6379"}}

		noopMP := metrics.NewNoopMetricsProvider()
		h, histErr := noopMP.NewFloat64Histogram("test")
		must.NoError(t, histErr)

		mp := &ProviderMock{
			NewInt64CounterFunc: func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metrics.Int64CounterForTest(t, "x"), nil
			},
			NewFloat64HistogramFunc: func(metricName string, _ ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
				test.EqOp(t, name+"_cache_latency_ms", metricName)
				return h, errors.New("histogram error")
			},
		}

		c, err := NewRedisCache[example](cfg, time.Minute, nil, nil, mp, nil)
		test.Error(t, err)
		test.Nil(t, c)
		test.SliceLen(t, 5, mp.NewInt64CounterCalls())
		test.SliceLen(t, 1, mp.NewFloat64HistogramCalls())
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

		cb.CannotProceedFunc = func() bool { return false }
		cb.SucceededFunc = func() {}

		client.GetFunc = func(_ context.Context, key string) *redis.StringCmd {
			test.EqOp(t, exampleKey, key)
			cmd := redis.NewStringCmd(ctx)
			cmd.SetVal(encoded)
			return cmd
		}

		actual, err := impl.Get(ctx, exampleKey)
		test.NoError(t, err)
		test.Eq(t, expected, actual)

		test.SliceLen(t, 1, client.GetCalls())
		test.SliceLen(t, 1, cb.CannotProceedCalls())
		test.SliceLen(t, 1, cb.SucceededCalls())
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return true }

		actual, err := impl.Get(ctx, exampleKey)
		test.ErrorIs(t, err, cache.ErrNotFound)
		test.Nil(t, actual)

		test.SliceLen(t, 1, cb.CannotProceedCalls())
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }
		cb.FailedFunc = func() {}

		client.GetFunc = func(_ context.Context, key string) *redis.StringCmd {
			test.EqOp(t, exampleKey, key)
			cmd := redis.NewStringCmd(ctx)
			cmd.SetErr(errors.New("redis error"))
			return cmd
		}

		actual, err := impl.Get(ctx, exampleKey)
		test.Error(t, err)
		test.Nil(t, actual)

		test.SliceLen(t, 1, client.GetCalls())
		test.SliceLen(t, 1, cb.FailedCalls())
	})

	T.Run("with decode error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }

		client.GetFunc = func(_ context.Context, key string) *redis.StringCmd {
			test.EqOp(t, exampleKey, key)
			cmd := redis.NewStringCmd(ctx)
			cmd.SetVal("not valid gob data")
			return cmd
		}

		actual, err := impl.Get(ctx, exampleKey)
		test.Error(t, err)
		test.Nil(t, actual)

		test.SliceLen(t, 1, client.GetCalls())
	})
}

func Test_redisCacheImpl_Set_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }
		cb.SucceededFunc = func() {}

		client.SetFunc = func(_ context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
			test.EqOp(t, exampleKey, key)
			test.EqOp(t, time.Minute, expiration)
			_, isString := value.(string)
			test.True(t, isString)
			cmd := redis.NewStatusCmd(ctx)
			cmd.SetVal("OK")
			return cmd
		}

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		test.NoError(t, err)

		test.SliceLen(t, 1, client.SetCalls())
		test.SliceLen(t, 1, cb.SucceededCalls())
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return true }

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		test.NoError(t, err)

		test.SliceLen(t, 1, cb.CannotProceedCalls())
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }
		cb.FailedFunc = func() {}

		client.SetFunc = func(_ context.Context, key string, _ any, _ time.Duration) *redis.StatusCmd {
			test.EqOp(t, exampleKey, key)
			cmd := redis.NewStatusCmd(ctx)
			cmd.SetErr(errors.New("redis error"))
			return cmd
		}

		err := impl.Set(ctx, exampleKey, &example{Name: t.Name()})
		test.Error(t, err)

		test.SliceLen(t, 1, client.SetCalls())
		test.SliceLen(t, 1, cb.FailedCalls())
	})
}

func Test_redisCacheImpl_Delete_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }
		cb.SucceededFunc = func() {}

		client.DelFunc = func(_ context.Context, keys ...string) *redis.IntCmd {
			test.Eq(t, []string{exampleKey}, keys)
			cmd := redis.NewIntCmd(ctx)
			cmd.SetVal(1)
			return cmd
		}

		err := impl.Delete(ctx, exampleKey)
		test.NoError(t, err)

		test.SliceLen(t, 1, client.DelCalls())
		test.SliceLen(t, 1, cb.SucceededCalls())
	})

	T.Run("when circuit breaker cannot proceed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, _, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return true }

		err := impl.Delete(ctx, exampleKey)
		test.NoError(t, err)

		test.SliceLen(t, 1, cb.CannotProceedCalls())
	})

	T.Run("with redis error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, cb := buildTestImpl(t)

		cb.CannotProceedFunc = func() bool { return false }
		cb.FailedFunc = func() {}

		client.DelFunc = func(_ context.Context, _ ...string) *redis.IntCmd {
			cmd := redis.NewIntCmd(ctx)
			cmd.SetErr(errors.New("redis error"))
			return cmd
		}

		err := impl.Delete(ctx, exampleKey)
		test.Error(t, err)

		test.SliceLen(t, 1, client.DelCalls())
		test.SliceLen(t, 1, cb.FailedCalls())
	})
}

func Test_redisCacheImpl_Ping_Unit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, _ := buildTestImpl(t)

		client.PingFunc = func(_ context.Context) *redis.StatusCmd {
			cmd := redis.NewStatusCmd(ctx)
			cmd.SetVal("PONG")
			return cmd
		}

		test.NoError(t, impl.Ping(ctx))
		test.SliceLen(t, 1, client.PingCalls())
	})

	T.Run("with error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		impl, client, _ := buildTestImpl(t)

		client.PingFunc = func(_ context.Context) *redis.StatusCmd {
			cmd := redis.NewStatusCmd(ctx)
			cmd.SetErr(errors.New("connection refused"))
			return cmd
		}

		test.Error(t, impl.Ping(ctx))
		test.SliceLen(t, 1, client.PingCalls())
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
		test.NotNil(t, c)
	})

	T.Run("with multiple addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			QueueAddresses: []string{"localhost:6379", "localhost:6380"},
			Username:       "user",
			Password:       "pass",
		}

		c := buildRedisClient(cfg)
		test.NotNil(t, c)
	})

	T.Run("with no addresses", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			QueueAddresses: []string{},
		}

		c := buildRedisClient(cfg)
		test.Nil(t, c)
	})
}
