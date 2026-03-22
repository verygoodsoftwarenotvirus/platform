package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v2/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildRedisBackedConsumer builds a Redis container-backed messagequeue.Consumer.
func buildRedisBackedConsumer(t *testing.T, cfg *Config, topic string, handlerFunc func(context.Context, []byte) error) messagequeue.Consumer {
	t.Helper()

	provider := ProvideRedisConsumerProvider(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		*cfg,
	)

	consumer, err := provider.ProvideConsumer(t.Context(), topic, handlerFunc)
	require.NoError(t, err)

	return consumer
}

func Test_redisConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg, containerShutdown, err := BuildContainerBackedRedisConfigForTest(t)
		if err != nil {
			t.Skipf("Skipping test due to container setup failure: %v", err)
		}
		defer func() {
			if containerShutdown != nil {
				assert.NoError(t, containerShutdown(ctx))
			}
		}()

		hf := func(context.Context, []byte) error {
			return nil
		}

		consumer := buildRedisBackedConsumer(t, cfg, t.Name(), hf)
		require.NotNil(t, consumer)

		stopChan := make(chan bool)
		errorsChan := make(chan error)
		go consumer.Consume(ctx, stopChan, errorsChan)

		publisher := buildRedisBackedPublisher(t, cfg, t.Name())
		require.NoError(t, publisher.Publish(ctx, []byte("blah")))

		<-time.After(time.Second)
		stopChan <- true
	})

	T.Run("with error handling message", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg, containerShutdown, err := BuildContainerBackedRedisConfigForTest(t)
		if err != nil {
			t.Skipf("Skipping test due to container setup failure: %v", err)
		}
		defer func() {
			if containerShutdown != nil {
				assert.NoError(t, containerShutdown(ctx))
			}
		}()

		anticipatedError := errors.New("blah")
		hf := func(context.Context, []byte) error {
			return anticipatedError
		}

		consumer := buildRedisBackedConsumer(t, cfg, t.Name(), hf)
		require.NotNil(t, consumer)

		stopChan := make(chan bool)
		errorsChan := make(chan error)
		go consumer.Consume(ctx, stopChan, errorsChan)

		publisher := buildRedisBackedPublisher(t, cfg, t.Name())
		require.NoError(t, publisher.Publish(ctx, []byte("blah")))

		receivedErr := <-errorsChan
		assert.Error(t, receivedErr)
		assert.Equal(t, anticipatedError, receivedErr)

		stopChan <- true
	})
}

func Test_consumerProvider_ProvideConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}

		conPro := ProvideRedisConsumerProvider(logger, tracing.NewNoopTracerProvider(), cfg)
		require.NotNil(t, conPro)

		ctx := t.Context()

		actual, err := conPro.ProvideConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("hitting cache", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}

		conPro := ProvideRedisConsumerProvider(logger, tracing.NewNoopTracerProvider(), cfg)
		require.NotNil(t, conPro)

		ctx := t.Context()

		actual, err := conPro.ProvideConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		actual, err = conPro.ProvideConsumer(ctx, t.Name(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}
