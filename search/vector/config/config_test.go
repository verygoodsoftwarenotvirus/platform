package vectorsearchcfg

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	circuitbreakingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/vector/pgvector"
	"github.com/verygoodsoftwarenotvirus/platform/v5/search/vector/qdrant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
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

	T.Run("pgvector provider with nil db returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: PgvectorProvider,
			Pgvector: &pgvector.Config{
				Dimension: 3,
				Metric:    vectorsearch.DistanceCosine,
			},
		}

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			cfg,
			nil,
			"idx",
		)
		assert.Error(t, err)
		assert.Nil(t, idx)
	})

	T.Run("qdrant provider via httptest server", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/collections/stub"):
				w.WriteHeader(http.StatusNotFound)
			case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/collections/stub"):
				_, _ = json.Marshal(r.Body)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":true,"status":"ok","time":0}`))
			default:
				http.Error(w, "unexpected", http.StatusBadRequest)
			}
		}))
		t.Cleanup(srv.Close)

		cfg := &Config{
			Provider: QdrantProvider,
			Qdrant: &qdrant.Config{
				BaseURL:   srv.URL,
				Dimension: 3,
				Metric:    vectorsearch.DistanceCosine,
				Timeout:   time.Second,
			},
		}

		idx, err := ProvideIndex[testStruct](
			t.Context(),
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			cfg,
			nil,
			"stub",
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
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
		mp.On("NewInt64Counter", "test-breaker_circuit_breaker_tripped", []metric.Int64CounterOption(nil)).
			Return(&mockmetrics.Int64Counter{}, errors.New("counter init failure"))

		idx, err := ProvideIndex[testStruct](
			ctx,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			cfg,
			nil,
			"idx",
		)
		assert.Error(t, err)
		assert.Nil(t, idx)
		assert.Contains(t, err.Error(), "circuit breaker")
		mp.AssertExpectations(t)
	})
}
