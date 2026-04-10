package pubsub

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

func TestBuildPubSubPublisher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		publisher := buildPubSubPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), nil, "test-topic")
		require.NotNil(t, publisher)
	})

	T.Run("panics when first NewInt64Counter fails", func(t *testing.T) {
		t.Parallel()

		mp := &metrics.MockProvider{}
		mp.On("NewInt64Counter", "t_published", mock.Anything).Return(metricnoop.Int64Counter{}, errors.New("forced error"))

		assert.Panics(t, func() {
			buildPubSubPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
		})

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("panics when second NewInt64Counter fails", func(t *testing.T) {
		t.Parallel()

		mp := &metrics.MockProvider{}
		mp.On("NewInt64Counter", "t_published", mock.Anything).Return(metricnoop.Int64Counter{}, nil)
		mp.On("NewInt64Counter", "t_publish_errors", mock.Anything).Return(metricnoop.Int64Counter{}, errors.New("forced error"))

		assert.Panics(t, func() {
			buildPubSubPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
		})

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("panics when NewFloat64Histogram fails", func(t *testing.T) {
		t.Parallel()

		mp := &metrics.MockProvider{}
		mp.On("NewInt64Counter", "t_published", mock.Anything).Return(metricnoop.Int64Counter{}, nil)
		mp.On("NewInt64Counter", "t_publish_errors", mock.Anything).Return(metricnoop.Int64Counter{}, nil)
		mp.On("NewFloat64Histogram", "t_publish_latency_ms", mock.Anything).Return(metricnoop.Float64Histogram{}, errors.New("forced error"))

		assert.Panics(t, func() {
			buildPubSubPublisher(logging.NewNoopLogger(), nil, tracing.NewNoopTracerProvider(), mp, "t")
		})

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestProvidePubSubPublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		provider := ProvidePubSubPublisherProvider(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, nil, "test-project")
		require.NotNil(t, provider)
	})
}

func TestPublisherProvider_Ping(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		p := &publisherProvider{}
		assert.NoError(t, p.Ping(t.Context()))
	})
}

func TestPublisherProvider_qualifyTopicName(T *testing.T) {
	T.Parallel()

	T.Run("already qualified", func(t *testing.T) {
		t.Parallel()

		p := &publisherProvider{projectID: "my-project"}
		result := p.qualifyTopicName("projects/my-project/topics/my-topic")
		assert.Equal(t, "projects/my-project/topics/my-topic", result)
	})

	T.Run("unqualified", func(t *testing.T) {
		t.Parallel()

		p := &publisherProvider{projectID: "my-project"}
		result := p.qualifyTopicName("my-topic")
		assert.Equal(t, "projects/my-project/topics/my-topic", result)
	})
}

func TestPublisherProvider_ProvidePublisher(T *testing.T) {
	T.Parallel()

	T.Run("with empty topic", func(t *testing.T) {
		t.Parallel()

		provider := ProvidePubSubPublisherProvider(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, nil, "test-project")

		pub, err := provider.ProvidePublisher(t.Context(), "")
		assert.Nil(t, pub)
		assert.ErrorIs(t, err, messagequeue.ErrEmptyTopicName)
	})
}
