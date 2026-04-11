package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/encoding"
	mockencoding "github.com/verygoodsoftwarenotvirus/platform/v5/encoding/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
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

type mockKafkaWriter struct {
	writeMessagesFunc func(ctx context.Context, msgs ...kafka.Message) error
	closeFunc         func() error
	writeCalls        int
	closeCalls        int
}

func (m *mockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.writeCalls++
	if m.writeMessagesFunc == nil {
		return nil
	}
	return m.writeMessagesFunc(ctx, msgs...)
}

func (m *mockKafkaWriter) Close() error {
	m.closeCalls++
	if m.closeFunc == nil {
		return nil
	}
	return m.closeFunc()
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
		writer.closeFunc = func() error { return nil }

		pub.Stop()

		assert.Equal(t, 1, writer.closeCalls)
	})

	T.Run("with close error", func(t *testing.T) {
		t.Parallel()

		pub, writer := buildTestPublisher(t)
		writer.closeFunc = func() error { return errors.New("close failed") }

		pub.Stop()

		assert.Equal(t, 1, writer.closeCalls)
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

		writer.writeMessagesFunc = func(_ context.Context, _ ...kafka.Message) error { return nil }

		err := pub.Publish(ctx, inputData)
		assert.NoError(t, err)

		assert.Equal(t, 1, writer.writeCalls)
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

		writer.writeMessagesFunc = func(_ context.Context, _ ...kafka.Message) error { return errors.New("write failed") }

		err := pub.Publish(ctx, inputData)
		assert.Error(t, err)

		assert.Equal(t, 1, writer.writeCalls)
	})

	T.Run("with mock encoder error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pub, _ := buildTestPublisher(t)

		enc := &mockencoding.ClientEncoderMock{
			EncodeFunc: func(_ context.Context, _ io.Writer, _ any) error {
				return errors.New("encode failed")
			},
		}
		pub.encoder = enc

		err := pub.Publish(ctx, "something")
		assert.Error(t, err)

		test.SliceLen(t, 1, enc.EncodeCalls())
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

		writer.writeMessagesFunc = func(_ context.Context, _ ...kafka.Message) error { return nil }

		pub.PublishAsync(ctx, inputData)

		assert.Equal(t, 1, writer.writeCalls)
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

		writer.writeMessagesFunc = func(_ context.Context, _ ...kafka.Message) error { return errors.New("write failed") }

		pub.PublishAsync(ctx, inputData)

		assert.Equal(t, 1, writer.writeCalls)
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

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metrics.Int64CounterForTest(t, "x"), errors.New("counter error")
			},
		}

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

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating publish error counter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.ProviderMock{}
		mp.NewInt64CounterFunc = func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
			if len(mp.NewInt64CounterCalls()) >= 2 {
				return metrics.Int64CounterForTest(t, "x"), errors.New("counter error")
			}
			return metrics.Int64CounterForTest(t, "x"), nil
		}

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

		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metrics.Int64CounterForTest(t, "x"), nil
			},
			NewFloat64HistogramFunc: func(_ string, _ ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
				return &metrics.Float64HistogramImpl{}, errors.New("histogram error")
			},
		}

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

		test.SliceLen(t, 1, mp.NewFloat64HistogramCalls())
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
		mw := &mockKafkaWriter{
			closeFunc: func() error { return nil },
		}

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

		assert.Equal(t, 1, mw.closeCalls)
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
