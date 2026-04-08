package qdrant

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/identifiers"
	vectorsearch "github.com/verygoodsoftwarenotvirus/platform/v5/search/vector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const qdrantImage = "qdrant/qdrant:v1.13.0"

var runningContainerTests = strings.ToLower(os.Getenv("RUN_CONTAINER_TESTS")) == "true"

type doc struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// --------- unit tests (no container required) ---------

func TestMetricToDistance(T *testing.T) {
	T.Parallel()

	cases := []struct {
		metric vectorsearch.DistanceMetric
		want   string
	}{
		{vectorsearch.DistanceCosine, "Cosine"},
		{vectorsearch.DistanceDotProduct, "Dot"},
		{vectorsearch.DistanceEuclidean, "Euclid"},
	}
	for _, c := range cases {
		T.Run(string(c.metric), func(t *testing.T) {
			t.Parallel()
			got, err := metricToDistance(c.metric)
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := metricToDistance("nonsense")
		require.ErrorIs(t, err, vectorsearch.ErrInvalidMetric)
	})
}

func TestStringifyID(T *testing.T) {
	T.Parallel()

	T.Run("string", func(t *testing.T) {
		t.Parallel()
		s, err := stringifyID("abc")
		require.NoError(t, err)
		assert.Equal(t, "abc", s)
	})

	T.Run("float", func(t *testing.T) {
		t.Parallel()
		s, err := stringifyID(float64(42))
		require.NoError(t, err)
		assert.Equal(t, "42", s)
	})

	T.Run("number", func(t *testing.T) {
		t.Parallel()
		s, err := stringifyID(json.Number("17"))
		require.NoError(t, err)
		assert.Equal(t, "17", s)
	})

	T.Run("unsupported", func(t *testing.T) {
		t.Parallel()
		_, err := stringifyID(true)
		require.Error(t, err)
	})
}

func TestUnmarshalPayload(T *testing.T) {
	T.Parallel()

	T.Run("nil round-trip", func(t *testing.T) {
		t.Parallel()
		out, err := unmarshalPayload[doc](nil)
		require.NoError(t, err)
		assert.Nil(t, out)
	})

	T.Run("populated", func(t *testing.T) {
		t.Parallel()
		out, err := unmarshalPayload[doc](json.RawMessage(`{"kind":"doc","title":"hi"}`))
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "doc", out.Kind)
		assert.Equal(t, "hi", out.Title)
	})

	T.Run("null", func(t *testing.T) {
		t.Parallel()
		out, err := unmarshalPayload[doc](json.RawMessage("null"))
		require.NoError(t, err)
		assert.Nil(t, out)
	})
}

func TestProvideIndex(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, nil, "test", cbnoop.NewCircuitBreaker())
		require.ErrorIs(t, err, vectorsearch.ErrNilConfig)
	})

	T.Run("empty collection", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{
			BaseURL:   "http://example",
			Dimension: 3,
			Metric:    vectorsearch.DistanceCosine,
		}, "", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})

	T.Run("invalid metric", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{
			BaseURL:   "http://example",
			Dimension: 3,
			Metric:    "weird",
		}, "test", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})

	T.Run("invalid dimension", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{
			BaseURL:   "http://example",
			Dimension: 0,
			Metric:    vectorsearch.DistanceCosine,
		}, "test", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})
}

// httptest-based test verifies the request shape we send to qdrant without needing a real qdrant.
func TestProvideIndex_StubsCollectionCreation(T *testing.T) {
	T.Parallel()

	var (
		gotMethod string
		gotPath   string
		gotBody   map[string]any
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/collections/stub"):
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/collections/stub"):
			gotMethod = r.Method
			gotPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":true,"status":"ok","time":0}`))
		default:
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	idx, err := ProvideIndex[doc](
		T.Context(),
		nil, nil, nil,
		&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
		"stub",
		cbnoop.NewCircuitBreaker(),
	)
	require.NoError(T, err)
	require.NotNil(T, idx)

	assert.Equal(T, http.MethodPut, gotMethod)
	assert.True(T, strings.HasSuffix(gotPath, "/collections/stub"))
	require.NotNil(T, gotBody)
	vectors, ok := gotBody["vectors"].(map[string]any)
	require.True(T, ok)
	assert.Equal(T, float64(3), vectors["size"])
	assert.Equal(T, "Cosine", vectors["distance"])
}

// --------- container-backed integration tests ---------

func buildContainerBackedQdrant(t *testing.T) (cfg *Config, shutdown func(context.Context) error) {
	t.Helper()

	ctx := t.Context()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        qdrantImage,
			ExposedPorts: []string{"6333/tcp"},
			WaitingFor:   wait.ForHTTP("/readyz").WithPort("6333/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, container)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6333/tcp")
	require.NoError(t, err)

	cfg = &Config{
		BaseURL:   "http://" + net.JoinHostPort(host, port.Port()),
		Dimension: 3,
		Metric:    vectorsearch.DistanceCosine,
		Timeout:   30 * time.Second,
	}
	return cfg, func(ctx context.Context) error { return container.Terminate(ctx) }
}

func TestQdrantIndex_Container(T *testing.T) {
	T.Parallel()

	if !runningContainerTests {
		T.SkipNow()
	}

	cfg, shutdown := buildContainerBackedQdrant(T)
	T.Cleanup(func() { _ = shutdown(context.Background()) })

	provide := func(t *testing.T, name string) vectorsearch.Index[doc] {
		t.Helper()
		idx, err := ProvideIndex[doc](t.Context(), nil, nil, nil, cfg, name, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		return idx
	}

	T.Run("Upsert and Query roundtrip", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "rt_"+identifiers.New())

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "11111111-1111-1111-1111-111111111111", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "doc", Title: "alpha"}},
			vectorsearch.Vector[doc]{ID: "22222222-2222-2222-2222-222222222222", Embedding: []float32{0, 1, 0}, Metadata: &doc{Kind: "doc", Title: "beta"}},
			vectorsearch.Vector[doc]{ID: "33333333-3333-3333-3333-333333333333", Embedding: []float32{0, 0, 1}, Metadata: &doc{Kind: "doc", Title: "gamma"}},
		))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 3})
		require.NoError(t, err)
		require.Len(t, results, 3)
		assert.Equal(t, "11111111-1111-1111-1111-111111111111", results[0].ID)
		require.NotNil(t, results[0].Metadata)
		assert.Equal(t, "alpha", results[0].Metadata.Title)
	})

	T.Run("TopK is respected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "topk_"+identifiers.New())

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "11111111-aaaa-aaaa-aaaa-111111111111", Embedding: []float32{1, 0, 0}},
			vectorsearch.Vector[doc]{ID: "22222222-aaaa-aaaa-aaaa-222222222222", Embedding: []float32{0, 1, 0}},
			vectorsearch.Vector[doc]{ID: "33333333-aaaa-aaaa-aaaa-333333333333", Embedding: []float32{0, 0, 1}},
		))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	T.Run("filter is applied", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "filt_"+identifiers.New())

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "11111111-bbbb-bbbb-bbbb-111111111111", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "doc"}},
			vectorsearch.Vector[doc]{ID: "22222222-bbbb-bbbb-bbbb-222222222222", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "image"}},
		))

		filter := map[string]any{
			"must": []any{
				map[string]any{
					"key":   "kind",
					"match": map[string]any{"value": "doc"},
				},
			},
		}
		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 10, Filter: filter})
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.NotNil(t, results[0].Metadata)
		assert.Equal(t, "doc", results[0].Metadata.Kind)
	})

	T.Run("Query rejects empty embedding", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "emb_"+identifiers.New())

		_, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: nil, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrEmptyEmbedding)
	})

	T.Run("Query rejects wrong dimension", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "dim_"+identifiers.New())

		_, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0}, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrDimensionMismatch)
	})

	T.Run("Delete removes specific points", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "del_"+identifiers.New())

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "11111111-cccc-cccc-cccc-111111111111", Embedding: []float32{1, 0, 0}},
			vectorsearch.Vector[doc]{ID: "22222222-cccc-cccc-cccc-222222222222", Embedding: []float32{0, 1, 0}},
		))

		require.NoError(t, idx.Delete(ctx, "11111111-cccc-cccc-cccc-111111111111"))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{0, 1, 0}, TopK: 10})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "22222222-cccc-cccc-cccc-222222222222", results[0].ID)
	})

	T.Run("Wipe drops and recreates", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		idx := provide(t, "wipe_"+identifiers.New())

		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "11111111-dddd-dddd-dddd-111111111111", Embedding: []float32{1, 0, 0}},
		))
		require.NoError(t, idx.Wipe(ctx))

		results, err := idx.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 10})
		require.NoError(t, err)
		assert.Empty(t, results)

		// Confirm the collection still accepts writes after wipe.
		require.NoError(t, idx.Upsert(ctx,
			vectorsearch.Vector[doc]{ID: "22222222-dddd-dddd-dddd-222222222222", Embedding: []float32{1, 0, 0}},
		))
	})

	T.Run("ProvideIndex is idempotent", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		name := "idem_" + identifiers.New()
		idx1, err := ProvideIndex[doc](ctx, nil, nil, nil, cfg, name, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)
		idx2, err := ProvideIndex[doc](ctx, nil, nil, nil, cfg, name, cbnoop.NewCircuitBreaker())
		require.NoError(t, err)

		require.NoError(t, idx1.Upsert(ctx, vectorsearch.Vector[doc]{ID: "11111111-eeee-eeee-eeee-111111111111", Embedding: []float32{1, 0, 0}}))

		results, err := idx2.Query(ctx, vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 1})
		require.NoError(t, err)
		require.Len(t, results, 1)
	})
}
