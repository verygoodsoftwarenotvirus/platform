package vectorsearchcfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/vector/pgvector"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/vector/qdrant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("pgvector provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: PgvectorProvider,
			Pgvector: &pgvector.Config{
				Dimension: 3,
				Metric:    vectorsearch.DistanceCosine,
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("qdrant provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: QdrantProvider,
			Qdrant: &qdrant.Config{
				BaseURL:   "http://localhost:6333",
				Dimension: 3,
				Metric:    vectorsearch.DistanceCosine,
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: "made-up"}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("pgvector provider without config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: PgvectorProvider}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("qdrant provider without config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: QdrantProvider}
		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("empty provider is valid (defaults to noop)", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestConfig_ProvideIndex(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			nil,
			nil,
			"idx",
		)
		assert.ErrorIs(t, err, vectorsearch.ErrNilConfig)
		assert.Nil(t, idx)
	})

	T.Run("unknown provider returns noop", func(t *testing.T) {
		t.Parallel()

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			&Config{Provider: "unknown"},
			nil,
			"idx",
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
		assert.NoError(t, idx.Wipe(t.Context()))
	})

	T.Run("empty provider returns noop", func(t *testing.T) {
		t.Parallel()

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			&Config{},
			nil,
			"idx",
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
	})

	T.Run("provider with whitespace returns noop", func(t *testing.T) {
		t.Parallel()

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			&Config{Provider: "   "},
			nil,
			"idx",
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
	})
}
