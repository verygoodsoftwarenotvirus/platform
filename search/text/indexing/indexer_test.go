package indexing

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue/config"
	mockpublishers "github.com/verygoodsoftwarenotvirus/platform/v5/messagequeue/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/reflection"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

var testQueuesConfig = &msgconfig.QueuesConfig{SearchIndexRequestsTopicName: "search_index_requests"}

func TestNewIndexScheduler(T *testing.T) {
	T.Parallel()

	T.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return []string{"id1", "id2"}, nil
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		assert.NoError(t, err)
		assert.NotNil(t, scheduler)
		assert.Equal(t, []string{"test_type"}, scheduler.allIndexTypes)
		assert.Len(t, scheduler.indexFunctions, 1)

		mock.AssertExpectationsForObjects(t, metricsProvider, messageQueueProvider)
	})

	T.Run("with nil index functions", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, nil)
		assert.NoError(t, err)
		assert.NotNil(t, scheduler)
		assert.Empty(t, scheduler.allIndexTypes)
		assert.NotNil(t, scheduler.indexFunctions)
		assert.Len(t, scheduler.indexFunctions, 0)

		mock.AssertExpectationsForObjects(t, metricsProvider, messageQueueProvider)
	})

	T.Run("metrics provider error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider to return error - need to return a valid interface and error
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, errors.New("metrics error"))

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, nil)
		assert.Error(t, err)
		assert.Nil(t, scheduler)
		assert.Contains(t, err.Error(), "metrics error")

		mock.AssertExpectationsForObjects(t, metricsProvider)
	})

	T.Run("message queue provider error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider to return error - need to return a valid interface and error
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, errors.New("message queue error"))

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, nil)
		assert.Error(t, err)
		assert.Nil(t, scheduler)
		assert.Contains(t, err.Error(), "message queue error")

		mock.AssertExpectationsForObjects(t, metricsProvider, messageQueueProvider)
	})
}

func TestIndexScheduler_IndexTypes(T *testing.T) {
	T.Parallel()

	T.Run("successful execution with results", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return []string{"id1", "id2", "id3"}, nil
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		// Mock publisher calls
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id1", IndexType: "test_type"}).Return(nil)
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id2", IndexType: "test_type"}).Return(nil)
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id3", IndexType: "test_type"}).Return(nil)

		// Mock metrics counter
		int64Counter.On(reflection.GetMethodName(int64Counter.Add), mock.Anything, int64(3), mock.Anything).Return()

		// Since we only have one index type, it will always be chosen
		err = scheduler.IndexTypes(ctx)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("successful execution with empty results", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function that returns empty results
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return []string{}, nil
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		// No publisher calls expected for empty results
		// But metrics counter is still called with 0
		int64Counter.On(reflection.GetMethodName(int64Counter.Add), mock.Anything, int64(0), mock.Anything).Return()

		err = scheduler.IndexTypes(ctx)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("function returns sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function that returns sql.ErrNoRows
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return nil, sql.ErrNoRows
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		// sql.ErrNoRows should be handled gracefully and return nil
		err = scheduler.IndexTypes(ctx)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("function returns other error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function that returns an error
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return nil, errors.New("database connection failed")
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		err = scheduler.IndexTypes(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection failed")

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("unknown index type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Create scheduler with empty index functions
		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, map[string]Function{})
		require.NoError(t, err)

		// This should not happen in normal operation since random.Element would return empty string
		// But we can test the error handling by directly calling with a non-existent type
		scheduler.allIndexTypes = []string{"unknown_type"}

		err = scheduler.IndexTypes(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown index type unknown_type")

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("partial publish failures", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return []string{"id1", "id2", "id3"}, nil
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		// Mock publisher calls - some succeed, some fail
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id1", IndexType: "test_type"}).Return(nil)
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id2", IndexType: "test_type"}).Return(errors.New("publish failed"))
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id3", IndexType: "test_type"}).Return(nil)

		// Mock metrics counter - should only count successful publishes
		int64Counter.On(reflection.GetMethodName(int64Counter.Add), mock.Anything, int64(2), mock.Anything).Return()

		err = scheduler.IndexTypes(ctx)
		assert.NoError(t, err) // Partial failures don't cause the method to return an error

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})

	T.Run("all publish failures", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := &mockmetrics.MetricsProvider{}
		messageQueueProvider := &mockpublishers.PublisherProvider{}

		// Mock metrics provider
		int64Counter := &mockmetrics.Int64Counter{}
		metricsProvider.On(reflection.GetMethodName(metricsProvider.NewInt64Counter), "indexer.handled_records", []metric.Int64CounterOption(nil)).Return(int64Counter, nil)

		// Mock message queue provider
		publisher := &mockpublishers.Publisher{}
		messageQueueProvider.On(reflection.GetMethodName(messageQueueProvider.ProvidePublisher), testQueuesConfig.SearchIndexRequestsTopicName).Return(publisher, nil)

		// Mock index function
		indexFunctions := map[string]Function{
			"test_type": func(ctx context.Context) ([]string, error) {
				return []string{"id1", "id2"}, nil
			},
		}

		scheduler, err := NewIndexScheduler(ctx, logger, tracerProvider, metricsProvider, messageQueueProvider, testQueuesConfig, indexFunctions)
		require.NoError(t, err)

		// Mock publisher calls - all fail
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id1", IndexType: "test_type"}).Return(errors.New("publish failed"))
		publisher.On(reflection.GetMethodName(publisher.Publish), mock.Anything, &textsearch.IndexRequest{RowID: "id2", IndexType: "test_type"}).Return(errors.New("publish failed"))

		// Mock metrics counter - should count 0 successful publishes
		int64Counter.On(reflection.GetMethodName(int64Counter.Add), mock.Anything, int64(0), mock.Anything).Return()

		err = scheduler.IndexTypes(ctx)
		assert.NoError(t, err) // Even all failures don't cause the method to return an error

		mock.AssertExpectationsForObjects(t, publisher, int64Counter)
	})
}
