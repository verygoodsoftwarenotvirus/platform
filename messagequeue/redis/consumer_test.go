package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

// buildRedisBackedConsumer builds a Redis container-backed messagequeue.Consumer.
func buildRedisBackedConsumer(t *testing.T, cfg *Config, topic string, handlerFunc func(context.Context, []byte) error) messagequeue.Consumer {
	t.Helper()

	provider := ProvideRedisConsumerProvider(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		nil,
		*cfg,
	)

	consumer, err := provider.ProvideConsumer(t.Context(), topic, handlerFunc)
	must.NoError(t, err)

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
				test.NoError(t, containerShutdown(ctx))
			}
		}()

		hf := func(context.Context, []byte) error {
			return nil
		}

		consumer := buildRedisBackedConsumer(t, cfg, t.Name(), hf)
		must.NotNil(t, consumer)

		stopChan := make(chan bool)
		errorsChan := make(chan error)
		go consumer.Consume(ctx, stopChan, errorsChan)

		publisher := buildRedisBackedPublisher(t, cfg, t.Name())
		must.NoError(t, publisher.Publish(ctx, []byte("blah")))

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
				test.NoError(t, containerShutdown(ctx))
			}
		}()

		anticipatedError := errors.New("blah")
		hf := func(context.Context, []byte) error {
			return anticipatedError
		}

		consumer := buildRedisBackedConsumer(t, cfg, t.Name(), hf)
		must.NotNil(t, consumer)

		stopChan := make(chan bool)
		errorsChan := make(chan error)
		go consumer.Consume(ctx, stopChan, errorsChan)

		publisher := buildRedisBackedPublisher(t, cfg, t.Name())
		must.NoError(t, publisher.Publish(ctx, []byte("blah")))

		receivedErr := <-errorsChan
		test.Error(t, receivedErr)
		test.ErrorIs(t, receivedErr, anticipatedError)

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

		conPro := ProvideRedisConsumerProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		must.NotNil(t, conPro)

		ctx := t.Context()

		actual, err := conPro.ProvideConsumer(ctx, t.Name(), nil)
		test.NoError(t, err)
		test.NotNil(t, actual)
	})

	T.Run("hitting cache", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}

		conPro := ProvideRedisConsumerProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		must.NotNil(t, conPro)

		ctx := t.Context()

		actual, err := conPro.ProvideConsumer(ctx, t.Name(), nil)
		test.NoError(t, err)
		test.NotNil(t, actual)

		actual, err = conPro.ProvideConsumer(ctx, t.Name(), nil)
		test.NoError(t, err)
		test.NotNil(t, actual)
	})

	T.Run("with empty topic", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := Config{
			QueueAddresses: []string{t.Name()},
		}

		conPro := ProvideRedisConsumerProvider(logger, tracing.NewNoopTracerProvider(), nil, cfg)
		must.NotNil(t, conPro)

		actual, err := conPro.ProvideConsumer(t.Context(), "", nil)
		test.Nil(t, actual)
		test.ErrorIs(t, err, ErrEmptyInputProvided)
	})
}

func Test_provideRedisConsumer(T *testing.T) {
	T.Parallel()

	T.Run("panics when NewInt64Counter fails", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(string, ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metricnoop.Int64Counter{}, errors.New("forced error")
			},
		}

		test.Panic(t, func() {
			provideRedisConsumer(t.Context(), logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp, nil, "t", nil)
		})
	})
}
