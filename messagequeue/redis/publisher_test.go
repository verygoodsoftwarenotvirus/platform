package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v2/testutils"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMessagePublisher struct {
	mock.Mock
}

// Publish implements the interface.
func (m *mockMessagePublisher) Publish(ctx context.Context, channel string, message any) *redis.IntCmd {
	return m.Called(ctx, channel, message).Get(0).(*redis.IntCmd)
}

// Close implements the interface.
func (m *mockMessagePublisher) Close() error {
	return m.Called().Error(0)
}

func buildRedisBackedPublisher(t *testing.T, cfg *Config, topic string) messagequeue.Publisher {
	t.Helper()

	ctx := t.Context()
	provider := ProvideRedisPublisherProvider(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
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
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), cfg)
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

		mmp := &mockMessagePublisher{}
		mmp.On(
			"Publish",
			testutils.ContextMatcher,
			actual.topic,
			fmt.Appendf(nil, `{"name":%q}%s`, t.Name(), string(byte(10))),
		).Return(&redis.IntCmd{})

		actual.publisher = mmp

		err = actual.Publish(ctx, inputData)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mmp)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), cfg)
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

func TestProvideRedisPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}
		actual := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), cfg)
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
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), cfg)
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
		provider := ProvideRedisPublisherProvider(logger, tracing.NewNoopTracerProvider(), cfg)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		actual, err = provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
