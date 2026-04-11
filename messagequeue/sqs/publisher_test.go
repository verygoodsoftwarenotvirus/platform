package sqs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

type mockMessagePublisher struct {
	sendMessageFunc  func(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	sendMessageCalls int
}

func (m *mockMessagePublisher) SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.sendMessageCalls++
	return m.sendMessageFunc(ctx, input, optFns...)
}

func Test_sqsPublisher_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{
			sendMessageFunc: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
				return &sqs.SendMessageOutput{}, nil
			},
		}

		actual.publisher = mmp

		err = actual.Publish(ctx, inputData)
		assert.NoError(t, err)
		assert.Equal(t, 1, mmp.sendMessageCalls)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
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

func Test_sqsPublisher_PublishAsync(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{
			sendMessageFunc: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
				return &sqs.SendMessageOutput{}, nil
			},
		}

		actual.publisher = mmp

		actual.PublishAsync(ctx, inputData)
		assert.Equal(t, 1, mmp.sendMessageCalls)
	})

	T.Run("with error encoding value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name json.Number `json:"name"`
		}{
			Name: json.Number(t.Name()),
		}

		actual.PublishAsync(ctx, inputData)
	})

	T.Run("with SendMessage error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		a, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, a)
		assert.NoError(t, err)

		actual, ok := a.(*sqsPublisher)
		require.True(t, ok)

		inputData := &struct {
			Name string `json:"name"`
		}{
			Name: t.Name(),
		}

		mmp := &mockMessagePublisher{
			sendMessageFunc: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
				return nil, errors.New("send failed")
			},
		}

		actual.publisher = mmp

		actual.PublishAsync(ctx, inputData)
		assert.Equal(t, 1, mmp.sendMessageCalls)
	})
}

func TestProvideSQSPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		actual := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		assert.NotNil(t, actual)
	})
}

func Test_publisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, t.Name())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
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

		provider := ProvideSQSPublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), nil)
		require.NotNil(t, provider)

		actual, err := provider.ProvidePublisher(ctx, "")
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
	})
}

func Test_provideSQSPublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		publisher := provideSQSPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), nil, "test-topic")
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
			provideSQSPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
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
			provideSQSPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
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
			provideSQSPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
		})
	})
}
