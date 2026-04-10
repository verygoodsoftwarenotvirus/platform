package indexing

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue/config"
	mockpublishers "github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/reflection"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	otelmetric "go.opentelemetry.io/otel/metric"
)

func TestRegisterIndexScheduler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		metricsProvider := &mockmetrics.MetricsProvider{}
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []otelmetric.Int64CounterOption(nil)).Return(int64Counter, nil)

		messageQueueProvider := &mockpublishers.PublisherProvider{}
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), "test_topic").Return(publisher, nil)

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue[metrics.Provider](i, metricsProvider)
		do.ProvideValue[messagequeue.PublisherProvider](i, messageQueueProvider)
		do.ProvideValue(i, &msgconfig.QueuesConfig{SearchIndexRequestsTopicName: "test_topic"})
		do.ProvideValue(i, map[string]Function{
			"test": func(ctx context.Context) ([]string, error) {
				return nil, nil
			},
		})

		RegisterIndexScheduler(i)

		scheduler, err := do.Invoke[*IndexScheduler](i)
		require.NoError(t, err)
		assert.NotNil(t, scheduler)

		mock.AssertExpectationsForObjects(t, metricsProvider, messageQueueProvider)
	})
}
