package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/encoding"
	mockencoding "github.com/verygoodsoftwarenotvirus/platform/v5/encoding/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
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

type mockKafkaWriter struct {
	mock.Mock
}

func (m *mockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.Called(ctx, msgs).Error(0)
}

func (m *mockKafkaWriter) Close() error {
	return m.Called().Error(0)
}

func buildTestPublisher(t *testing.T) (*kafkaPublisher, *mockKafkaWriter) {
	t.Helper()

	writer := &mockKafkaWriter{}

	mp := metrics.NewNoopMetricsProvider()

	publishedCounter, err := mp.NewInt64Counter("test_published")
	require.NoError(t, err)

	publishErrCounter, err := mp.NewInt64Counter("test_publish_errors")
	require.NoError(t, err)

	latencyHist, err := mp.NewFloat64Histogram("test_publish_latency_ms")
	require.NoError(t, err)

	pub := &kafkaPublisher{
		writer:            writer,
		encoder:           encoding.ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON),
		logger:            logging.NewNoopLogger(),
		tracer:            tracing.NewTracerForTest(t.Name()),
		publishedCounter:  publishedCounter,
		publishErrCounter: publishErrCounter,
		latencyHist:       latencyHist,
	}

	return pub, writer
}

func Test_kafkaPublisher_Stop(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		pub, writer := buildTestPublisher(t)
		writer.On("Close").Return(nil)

		pub.Stop()

		mock.AssertExpectationsForObjects(t, writer)
	})

	T.Run("with close error", func(t *testing.T) {
		t.Parallel()

		pub, writer := buildTestPublisher(t)
		writer.On("Close").Return(errors.New("close failed"))

		pub.Stop()

		mock.AssertExpectationsForObjects(t, writer)
	})
}

func Test_kafkaPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, writer := buildTestPublisher(t)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		writer.On("WriteMessages", testutils.ContextMatcher, mock.AnythingOfType("[]kafka.Message")).Return(nil)

		err := pub.Publish(ctx, inputData)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, writer)
	})

	T.Run("with encoding error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, _ := buildTestPublisher(t)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		err := pub.Publish(ctx, inputData)
		assert.Error(t, err)
	})

	T.Run("with write error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, writer := buildTestPublisher(t)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		writer.On("WriteMessages", testutils.ContextMatcher, mock.AnythingOfType("[]kafka.Message")).Return(errors.New("write failed"))

		err := pub.Publish(ctx, inputData)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, writer)
	})

	T.Run("with mock encoder error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, _ := buildTestPublisher(t)

		enc := &mockencoding.ClientEncoder{}
		enc.On("Encode", testutils.ContextMatcher, mock.Anything, mock.Anything).Return(errors.New("encode failed"))
		pub.encoder = enc

		err := pub.Publish(ctx, "something")
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, enc)
	})
}

func Test_kafkaPublisher_PublishAsync(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, writer := buildTestPublisher(t)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		done := make(chan struct{})
		writer.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(nil).Run(func(args mock.Arguments) {
			close(done)
		})

		pub.PublishAsync(ctx, inputData)

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for async write")
		}

		mock.AssertExpectationsForObjects(t, writer)
	})

	T.Run("with encoding error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, _ := buildTestPublisher(t)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		pub.PublishAsync(ctx, inputData)
	})

	T.Run("with write error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, writer := buildTestPublisher(t)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		done := make(chan struct{})
		writer.On("WriteMessages", mock.Anything, mock.AnythingOfType("[]kafka.Message")).Return(errors.New("write failed")).Run(func(args mock.Arguments) {
			close(done)
		})

		pub.PublishAsync(ctx, inputData)

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for async write")
		}

		mock.AssertExpectationsForObjects(t, writer)
	})
}

func TestProvideKafkaPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		actual := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		assert.NotNil(t, actual)
	})
}

func Test_publisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
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

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
		assert.Nil(t, actual)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		first, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NoError(t, err)
		assert.NotNil(t, first)

		second, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NoError(t, err)
		assert.NotNil(t, second)

		assert.Equal(t, first, second)
	})

	T.Run("with error creating published counter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.Anything, []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), errors.New("counter error"))

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error creating publish error counter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.Anything, []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), nil).Once()
		mp.On("NewInt64Counter", mock.Anything, []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), errors.New("counter error")).Once()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.Anything, []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest("x"), nil)
		mp.On("NewFloat64Histogram", mock.Anything, []metric.Float64HistogramOption(nil)).Return(&metrics.Float64HistogramImpl{}, errors.New("histogram error"))

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			cfg,
		)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func Test_publisherProvider_Ping(T *testing.T) {
	T.Parallel()

	T.Run("with unreachable broker", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:1"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		err := provider.Ping(ctx)
		assert.Error(t, err)
	})
}

func Test_publisherProvider_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		_, err := provider.ProvidePublisher(ctx, t.Name())
		require.NoError(t, err)

		pp, ok := provider.(*publisherProvider)
		require.True(t, ok)

		// Replace cached publisher with one using a mock writer so Close doesn't hit real Kafka
		mw := &mockKafkaWriter{}
		mw.On("Close").Return(nil)

		mp := metrics.NewNoopMetricsProvider()
		publishedCounter, _ := mp.NewInt64Counter("test_published")
		publishErrCounter, _ := mp.NewInt64Counter("test_publish_errors")
		latencyHist, _ := mp.NewFloat64Histogram("test_publish_latency_ms")

		pp.publisherCache[t.Name()] = &kafkaPublisher{
			writer:            mw,
			logger:            logging.NewNoopLogger(),
			tracer:            tracing.NewTracerForTest(t.Name()),
			publishedCounter:  publishedCounter,
			publishErrCounter: publishErrCounter,
			latencyHist:       latencyHist,
		}

		provider.Close()

		mock.AssertExpectationsForObjects(t, mw)
	})

	T.Run("with empty cache", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Brokers: []string{"localhost:9092"},
			GroupID: "test-group",
		}

		provider := ProvideKafkaPublisherProvider(
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			nil,
			cfg,
		)
		require.NotNil(t, provider)

		provider.Close()
	})
}
