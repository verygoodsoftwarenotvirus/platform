package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/segmentio/kafka-go"
	"github.com/shoenig/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

type mockKafkaReader struct {
	fetchMessageFunc   func(ctx context.Context) (kafka.Message, error)
	commitMessagesFunc func(ctx context.Context, msgs ...kafka.Message) error
	closeFunc          func() error
	fetchCalls         int
	commitCalls        int
}

func (m *mockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	m.fetchCalls++
	return m.fetchMessageFunc(ctx)
}

func (m *mockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.commitCalls++
	return m.commitMessagesFunc(ctx, msgs...)
}

func (m *mockKafkaReader) Close() error {
	if m.closeFunc == nil {
		return nil
	}
	return m.closeFunc()
}

func Test_kafkaConsumer_Consume(T *testing.T) {
	T.Parallel()

	T.Run("stops on context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				return kafka.Message{}, context.Canceled
			},
		}

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

		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				return kafka.Message{}, context.Canceled
			},
		}

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

		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				return kafka.Message{}, context.Canceled
			},
		}

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

		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				callCount++
				if callCount >= 2 {
					cancel()
				}
				return kafka.Message{}, fetchErr
			},
		}

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

		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				cancel()
				return kafka.Message{}, fetchErr
			},
		}

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

		var fetchCount int
		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				fetchCount++
				if fetchCount == 1 {
					return msg, nil
				}
				return kafka.Message{}, context.Canceled
			},
			commitMessagesFunc: func(_ context.Context, msgs ...kafka.Message) error {
				require.Len(t, msgs, 1)
				assert.Equal(t, msg, msgs[0])
				return nil
			},
		}

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
		assert.Equal(t, 1, reader.commitCalls)
	})

	T.Run("with handler error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		msg := kafka.Message{Value: []byte("test-message")}
		handlerErr := errors.New("handler failed")

		var fetchCount int
		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				fetchCount++
				if fetchCount == 1 {
					return msg, nil
				}
				return kafka.Message{}, context.Canceled
			},
		}

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

		var fetchCount int
		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				fetchCount++
				if fetchCount == 1 {
					return msg, nil
				}
				return kafka.Message{}, context.Canceled
			},
		}

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

		var fetchCount int
		reader := &mockKafkaReader{
			fetchMessageFunc: func(_ context.Context) (kafka.Message, error) {
				fetchCount++
				if fetchCount == 1 {
					return msg, nil
				}
				return kafka.Message{}, context.Canceled
			},
			commitMessagesFunc: func(_ context.Context, _ ...kafka.Message) error {
				return errors.New("commit failed")
			},
		}

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
		assert.Equal(t, 1, reader.commitCalls)
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

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metrics.Int64CounterForTest(t, "x"), errors.New("counter error")
			},
		}

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

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
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
