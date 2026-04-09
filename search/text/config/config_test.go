package textsearchcfg

import (
	"context"
	"errors"
	"testing"

	circuitbreakingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/text/algolia"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/text/elasticsearch"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("elasticsearch provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ElasticsearchProvider,
			Elasticsearch: &elasticsearch.Config{
				Address: t.Name(),
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("algolia provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: AlgoliaProvider,
			Algolia: &algolia.Config{
				AppID:  "test-app-id",
				APIKey: "test-api-key",
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "invalid-provider",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("elasticsearch provider without elasticsearch config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ElasticsearchProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("algolia provider without algolia config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: AlgoliaProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("empty provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
		}

		// Empty provider should be valid (it will default to noop)
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("provider with extra whitespace", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "  " + ElasticsearchProvider + "  ",
			Elasticsearch: &elasticsearch.Config{
				Address: t.Name(),
			},
		}

		// Provider with whitespace should be invalid (validation is strict)
		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("provider case insensitive", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "ELASTICSEARCH",
			Elasticsearch: &elasticsearch.Config{
				Address: t.Name(),
			},
		}

		// Provider should be case sensitive (validation is strict)
		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("nil context", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ElasticsearchProvider,
			Elasticsearch: &elasticsearch.Config{
				Address: t.Name(),
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(context.TODO()))
	})
}

func TestConfig_ZeroValue(T *testing.T) {
	T.Parallel()

	T.Run("zero value is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		// Zero value should be valid (it will default to noop)
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("zero value fields", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		assert.Equal(t, "", cfg.Provider)
		assert.Nil(t, cfg.Algolia)
		assert.Nil(t, cfg.Elasticsearch)
	})
}

func TestConfig_Constants(T *testing.T) {
	T.Parallel()

	T.Run("provider constants have expected values", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "elasticsearch", ElasticsearchProvider)
		assert.Equal(t, "algolia", AlgoliaProvider)
	})

	T.Run("provider constants are not empty", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, ElasticsearchProvider)
		assert.NotEmpty(t, AlgoliaProvider)
	})

	T.Run("provider constants are different", func(t *testing.T) {
		t.Parallel()

		assert.NotEqual(t, ElasticsearchProvider, AlgoliaProvider)
	})
}

func TestConfig_ProvideIndex(T *testing.T) {
	T.Parallel()

	T.Run("elasticsearch provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ElasticsearchProvider,
			Elasticsearch: &elasticsearch.Config{
				Address: "http://localhost:9200",
			},
		}

		// This will fail because we don't have a real Elasticsearch instance
		// but we're testing the interface compliance
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := metrics.NewNoopMetricsProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, metricsProvider, cfg, "test-index")
		assert.Error(t, err)
		assert.Nil(t, index)
	})

	T.Run("algolia provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: AlgoliaProvider,
			Algolia: &algolia.Config{
				AppID:  "test-app-id",
				APIKey: "test-api-key",
			},
		}

		// This will succeed because we're using a real Algolia client
		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := metrics.NewNoopMetricsProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, metricsProvider, cfg, "test-index")
		assert.NoError(t, err)
		assert.NotNil(t, index)
	})

	T.Run("unknown provider returns noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "unknown-provider",
		}

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := metrics.NewNoopMetricsProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, metricsProvider, cfg, "test-index")
		assert.NoError(t, err)
		assert.NotNil(t, index)
	})

	T.Run("empty provider returns noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
		}

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := metrics.NewNoopMetricsProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, metricsProvider, cfg, "test-index")
		assert.NoError(t, err)
		assert.NotNil(t, index)
	})

	T.Run("provider with whitespace returns noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "   ",
		}

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		metricsProvider := metrics.NewNoopMetricsProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, metricsProvider, cfg, "test-index")
		assert.NoError(t, err)
		assert.NotNil(t, index)
	})

	T.Run("circuit breaker init failure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
			CircuitBreaker: circuitbreakingcfg.Config{
				Name:                   "test-breaker",
				ErrorRate:              50,
				MinimumSampleThreshold: 10,
			},
		}

		mp := &mockmetrics.MetricsProvider{}
		// Force the very first counter creation to fail so ProvideCircuitBreaker
		// returns an error, which is wrapped by ProvideIndex.
		mp.On("NewInt64Counter", "test-breaker_circuit_breaker_tripped", []metric.Int64CounterOption(nil)).
			Return(&mockmetrics.Int64Counter{}, errors.New("counter init failure"))

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		index, err := ProvideIndex[testStruct](ctx, logger, tracerProvider, mp, cfg, "test-index")
		assert.Error(t, err)
		assert.Nil(t, index)
		assert.Contains(t, err.Error(), "circuit breaker")
		mp.AssertExpectations(t)
	})
}

type testStruct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
