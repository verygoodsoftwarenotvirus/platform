package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
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

func (m *mockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...any) *redis.Cmd {
	return m.Called(ctx, script, keys, args).Get(0).(*redis.Cmd)
}

func (m *mockRedisClient) Close() error {
	return m.Called().Error(0)
}

func buildTestRateLimiter(t *testing.T) (*rateLimiter, *mockRedisClient) {
	t.Helper()

	client := &mockRedisClient{}
	mp := metrics.NewNoopMetricsProvider()

	allowedCounter, err := mp.NewInt64Counter(redisName + "_allowed")
	require.NoError(t, err)

	rejectedCounter, err := mp.NewInt64Counter(redisName + "_rejected")
	require.NoError(t, err)

	errorCounter, err := mp.NewInt64Counter(redisName + "_errors")
	require.NoError(t, err)

	return &rateLimiter{
		client:          client,
		requestsPerSec:  10,
		allowedCounter:  allowedCounter,
		rejectedCounter: rejectedCounter,
		errorCounter:    errorCounter,
	}, client
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			Addresses: []string{"localhost:6379"},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty addresses", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			Addresses: []string{},
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

func TestNewRedisRateLimiter(T *testing.T) {
	T.Parallel()

	T.Run("with single address", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Addresses: []string{"localhost:6379"},
			Username:  "user",
			Password:  "pass",
		}

		rl, err := NewRedisRateLimiter(cfg, nil, 10)
		assert.NoError(t, err)
		assert.NotNil(t, rl)
	})

	T.Run("with multiple addresses", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Addresses: []string{"localhost:6379", "localhost:6380"},
			Username:  "user",
			Password:  "pass",
		}

		rl, err := NewRedisRateLimiter(cfg, nil, 10)
		assert.NoError(t, err)
		assert.NotNil(t, rl)
	})

	T.Run("with error creating allowed counter", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Addresses: []string{"localhost:6379"},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", redisName+"_allowed", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), errors.New("counter error"))

		rl, err := NewRedisRateLimiter(cfg, mp, 10)
		assert.Error(t, err)
		assert.Nil(t, rl)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating rejected counter", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Addresses: []string{"localhost:6379"},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", redisName+"_allowed", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), nil)
		mp.On("NewInt64Counter", redisName+"_rejected", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), errors.New("counter error"))

		rl, err := NewRedisRateLimiter(cfg, mp, 10)
		assert.Error(t, err)
		assert.Nil(t, rl)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Addresses: []string{"localhost:6379"},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", redisName+"_allowed", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), nil)
		mp.On("NewInt64Counter", redisName+"_rejected", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), nil)
		mp.On("NewInt64Counter", redisName+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), errors.New("counter error"))

		rl, err := NewRedisRateLimiter(cfg, mp, 10)
		assert.Error(t, err)
		assert.Nil(t, rl)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func Test_rateLimiter_Allow(T *testing.T) {
	T.Parallel()

	T.Run("allowed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rl, client := buildTestRateLimiter(t)

		cmd := redis.NewCmd(ctx)
		cmd.SetVal(int64(1))
		client.On("Eval", testutils.ContextMatcher, slidingWindowScript, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)

		allowed, err := rl.Allow(ctx, "test-key")
		assert.NoError(t, err)
		assert.True(t, allowed)

		mock.AssertExpectationsForObjects(t, client)
	})

	T.Run("rejected", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rl, client := buildTestRateLimiter(t)

		cmd := redis.NewCmd(ctx)
		cmd.SetVal(int64(0))
		client.On("Eval", testutils.ContextMatcher, slidingWindowScript, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)

		allowed, err := rl.Allow(ctx, "test-key")
		assert.NoError(t, err)
		assert.False(t, allowed)

		mock.AssertExpectationsForObjects(t, client)
	})

	T.Run("with eval error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rl, client := buildTestRateLimiter(t)

		cmd := redis.NewCmd(ctx)
		cmd.SetErr(errors.New("redis error"))
		client.On("Eval", testutils.ContextMatcher, slidingWindowScript, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]interface {}")).Return(cmd)

		allowed, err := rl.Allow(ctx, "test-key")
		assert.Error(t, err)
		assert.False(t, allowed)

		mock.AssertExpectationsForObjects(t, client)
	})
}

func Test_rateLimiter_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		rl, client := buildTestRateLimiter(t)
		client.On("Close").Return(nil)

		err := rl.Close()
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, client)
	})

	T.Run("with close error", func(t *testing.T) {
		t.Parallel()

		rl, client := buildTestRateLimiter(t)
		client.On("Close").Return(errors.New("close failed"))

		err := rl.Close()
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, client)
	})
}
