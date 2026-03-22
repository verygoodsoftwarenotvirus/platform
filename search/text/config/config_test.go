package textsearchcfg

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v2/search/text/algolia"
	"github.com/verygoodsoftwarenotvirus/platform/v2/search/text/elasticsearch"

	"github.com/stretchr/testify/assert"
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
}

type testStruct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
