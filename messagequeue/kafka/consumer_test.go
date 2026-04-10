package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/testutils"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

type mockKafkaReader struct {
	mock.Mock
}

func (m *mockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *mockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

func (m *mockKafkaReader) Close() error {
	return m.Called().Error(0)
}

func Test_kafkaConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("stops on context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: nil,
			handlerFunc: func(context.Context, []byte) error {
				return nil
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 1)

		cancel()
		c.Consume(ctx, stopChan, errs)
	})

	T.Run("stops on stop channel signal", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		reader := &mockKafkaReader{}

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: nil,
			handlerFunc: func(context.Context, []byte) error {
				return nil
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 1)

		stopChan <- true
		c.Consume(ctx, stopChan, errs)
	})

	T.Run("with nil stop channel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: nil,
			handlerFunc: func(context.Context, []byte) error {
				return nil
			},
		}

		errs := make(chan error, 1)

		cancel()
		c.Consume(ctx, nil, errs)
	})

	T.Run("with fetch error and context still alive", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		fetchErr := errors.New("fetch failed")
		callCount := 0

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, fetchErr).Run(func(args mock.Arguments) {
			callCount++
			if callCount >= 2 {
				cancel()
			}
		})

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: nil,
			handlerFunc: func(context.Context, []byte) error {
				return nil
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 10)

		c.Consume(ctx, stopChan, errs)

		select {
		case receivedErr := <-errs:
			assert.Equal(t, fetchErr, receivedErr)
		default:
			t.Error("expected an error on the errors channel")
		}
	})

	T.Run("with fetch error and nil errors channel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		fetchErr := errors.New("fetch failed")

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, fetchErr).Run(func(args mock.Arguments) {
			cancel()
		})

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: nil,
			handlerFunc: func(context.Context, []byte) error {
				return nil
			},
		}

		stopChan := make(chan bool, 1)

		c.Consume(ctx, stopChan, nil)
	})

	T.Run("with successful message handling", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		msg := kafka.Message{Value: []byte("test-message")}

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(msg, nil).Once()
		reader.On("CommitMessages", testutils.ContextMatcher, []kafka.Message{msg}).Return(nil).Once()
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		handlerCalled := false
		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: metrics.Int64CounterForTest(t, t.Name()),
			handlerFunc: func(_ context.Context, data []byte) error {
				handlerCalled = true
				assert.Equal(t, []byte("test-message"), data)
				cancel()
				return nil
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 10)

		c.Consume(ctx, stopChan, errs)
		assert.True(t, handlerCalled)

		mock.AssertExpectationsForObjects(t, reader)
	})

	T.Run("with handler error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		msg := kafka.Message{Value: []byte("test-message")}
		handlerErr := errors.New("handler failed")

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(msg, nil).Once()
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: metrics.Int64CounterForTest(t, t.Name()),
			handlerFunc: func(context.Context, []byte) error {
				cancel()
				return handlerErr
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 10)

		c.Consume(ctx, stopChan, errs)

		receivedErr := <-errs
		assert.Error(t, receivedErr)
		assert.Equal(t, handlerErr, receivedErr)
	})

	T.Run("with handler error and nil errors channel", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		msg := kafka.Message{Value: []byte("test-message")}

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(msg, nil).Once()
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: metrics.Int64CounterForTest(t, t.Name()),
			handlerFunc: func(context.Context, []byte) error {
				cancel()
				return errors.New("handler failed")
			},
		}

		stopChan := make(chan bool, 1)

		c.Consume(ctx, stopChan, nil)
	})

	T.Run("with commit error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		msg := kafka.Message{Value: []byte("test-message")}

		reader := &mockKafkaReader{}
		reader.On("FetchMessage", testutils.ContextMatcher).Return(msg, nil).Once()
		reader.On("CommitMessages", testutils.ContextMatcher, []kafka.Message{msg}).Return(errors.New("commit failed")).Once()
		reader.On("FetchMessage", testutils.ContextMatcher).Return(kafka.Message{}, context.Canceled).Maybe()

		c := &kafkaConsumer{
			reader:          reader,
			logger:          logging.NewNoopLogger(),
			tracer:          tracing.NewTracerForTest(t.Name()),
			consumedCounter: metrics.Int64CounterForTest(t, t.Name()),
			handlerFunc: func(context.Context, []byte) error {
				cancel()
				return nil
			},
		}

		stopChan := make(chan bool, 1)
		errs := make(chan error, 10)

		c.Consume(ctx, stopChan, errs)

		mock.AssertExpectationsForObjects(t, reader)
	})
}

func TestProvideKafkaConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		actual := ProvideKafkaConsumerProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		assert.NotNil(t, actual)
	})
}

func Test_consumerProvider_ProvideConsumer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaConsumerProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		hf := func(context.Context, []byte) error { return nil }

		actual, err := provider.ProvideConsumer(ctx, t.Name(), hf)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with empty topic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaConsumerProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvideConsumer(ctx, "", nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyInputProvided)
		assert.Nil(t, actual)
	})

	T.Run("with error creating consumed counter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.Anything, []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("counter error"))

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaConsumerProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			cfg,
		)
		require.NotNil(t, provider)

		hf := func(context.Context, []byte) error { return nil }

		actual, err := provider.ProvideConsumer(ctx, t.Name(), hf)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaConsumerProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		hf := func(context.Context, []byte) error { return nil }

		first, err := provider.ProvideConsumer(ctx, t.Name(), hf)
		assert.NoError(t, err)
		assert.NotNil(t, first)

		second, err := provider.ProvideConsumer(ctx, t.Name(), hf)
		assert.NoError(t, err)
		assert.NotNil(t, second)

		assert.Equal(t, first, second)
	})
}
