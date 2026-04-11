package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

type mockMessagePublisher struct {
	publishFunc func(ctx context.Context, channel string, message any) *redis.IntCmd
	closeFunc   func() error
	pingFunc    func(ctx context.Context) *redis.StatusCmd
	publishArgs []publishCall
}

type publishCall struct {
	ctx     context.Context
	message any
	channel string
}

func (m *mockMessagePublisher) Publish(ctx context.Context, channel string, message any) *redis.IntCmd {
	m.publishArgs = append(m.publishArgs, publishCall{ctx: ctx, channel: channel, message: message})
	return m.publishFunc(ctx, channel, message)
}

func (m *mockMessagePublisher) Close() error {
	return m.closeFunc()
}

func (m *mockMessagePublisher) Ping(ctx context.Context) *redis.StatusCmd {
	return m.pingFunc(ctx)
}

func buildRedisBackedPublisher(t *testing.T, cfg *Config, topic string) messagequeue.Publisher {
	t.Helper()

	ctx := t.Context()
	provider := ProvideRedisPublisherProvider(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		nil,
		*cfg,
	)

	publisher, err := provider.ProvidePublisher(ctx, topic)
	require.NoError(t, err)

	return publisher
}

func Test_redisPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{
			publishFunc: func(_ context.Context, _ string, _ any) *redis.IntCmd { return &redis.IntCmd{} },
		}

		actual.publisher = mmp

		err = actual.Publish(ctx, inputData)
		assert.NoError(t, err)

		require.Len(t, mmp.publishArgs, 1)
		assert.Equal(t, actual.topic, mmp.publishArgs[0].channel)
		assert.Equal(t, fmt.Appendf(nil, `{"name":%q}%s`, t.Name(), string(byte(10))), mmp.publishArgs[0].message)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		err = actual.Publish(ctx, inputData)
		assert.Error(t, err)
	})
}

func Test_redisPublisher_PublishAsync(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{
			publishFunc: func(_ context.Context, _ string, _ any) *redis.IntCmd { return &redis.IntCmd{} },
		}

		actual.publisher = mmp

		actual.PublishAsync(ctx, inputData)

		require.Len(t, mmp.publishArgs, 1)
		assert.Equal(t, actual.topic, mmp.publishArgs[0].channel)
		assert.Equal(t, fmt.Appendf(nil, `{"name":%q}%s`, t.Name(), string(byte(10))), mmp.publishArgs[0].message)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*redisPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		actual.PublishAsync(ctx, inputData)
	})
}

func TestProvideRedisPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		actual := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		assert.NotNil(t, actual)
	})
}

func Test_publisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		actual, err = provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with empty topic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, "")
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
	})
}

func Test_provideRedisPublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		publisher := provideRedisPublisher(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, nil, "test-topic")
		require.NotNil(t, publisher)
	})

	T.Run("panics when first NewInt64Counter fails", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(name string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				if name == "t_published" {
					return metricnoop.Int64Counter{}, errors.New("forced error")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", name)
				return nil, nil
			},
		}

		assert.Panics(t, func() {
			provideRedisPublisher(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, nil, "t")
		})
	})

	T.Run("panics when second NewInt64Counter fails", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(name string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch name {
				case "t_published":
					return metricnoop.Int64Counter{}, nil
				case "t_publish_errors":
					return metricnoop.Int64Counter{}, errors.New("forced error")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", name)
				return nil, nil
			},
		}

		assert.Panics(t, func() {
			provideRedisPublisher(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, nil, "t")
		})
	})

	T.Run("panics when NewFloat64Histogram fails", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(string, ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metricnoop.Int64Counter{}, nil
			},
			NewFloat64HistogramFunc: func(string, ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
				return metricnoop.Float64Histogram{}, errors.New("forced error")
			},
		}

		assert.Panics(t, func() {
			provideRedisPublisher(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, nil, "t")
		})
	})
}
