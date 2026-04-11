package qdrant

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	cbmock "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
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

// qdrantStub is a configurable httptest handler for qdrant REST endpoints.
type qdrantStub struct {
	pointsSearchBody       string
	collectionGetStatus    int
	collectionPutStatus    int
	pointsPutStatus        int
	pointsDeleteStatus     int
	collectionDeleteStatus int
	pointsSearchStatus     int
}

func (s *qdrantStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	okBody := `{"result":true,"status":"ok","time":0}`
	path := r.URL.Path

	switch {
	case r.Method == http.MethodGet && strings.Contains(path, "/collections/"):
		status := s.collectionGetStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(okBody))
		}

	case r.Method == http.MethodPut && strings.Contains(path, "/points"):
		status := s.pointsPutStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(okBody))
		}

	case r.Method == http.MethodPut && strings.Contains(path, "/collections/"):
		status := s.collectionPutStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(okBody))
		}

	case r.Method == http.MethodPost && strings.Contains(path, "/points/delete"):
		status := s.pointsDeleteStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(okBody))
		}

	case r.Method == http.MethodDelete && strings.Contains(path, "/collections/"):
		status := s.collectionDeleteStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(okBody))
		}

	case r.Method == http.MethodPost && strings.Contains(path, "/points/search"):
		status := s.pointsSearchStatus
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		body := s.pointsSearchBody
		if body == "" {
			body = `{"result":[]}`
		}
		_, _ = w.Write([]byte(body))

	default:
		http.Error(w, fmt.Sprintf("unexpected %s %s", r.Method, path), http.StatusBadRequest)
	}
}

// buildStubIndex creates an indexManager backed by an httptest server using the given stub.
// The server is closed when the test finishes.
func buildStubIndex(t *testing.T, stub *qdrantStub, cb circuitbreaking.CircuitBreaker) *indexManager[doc] {
	t.Helper()

	srv := httptest.NewServer(stub)
	t.Cleanup(srv.Close)

	if cb == nil {
		cb = cbnoop.NewCircuitBreaker()
	}

	idx, err := ProvideIndex[doc](
		t.Context(),
		nil, nil, nil,
		&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
		"test",
		cb,
	)
	require.NoError(t, err)

	return idx.(*indexManager[doc])
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

	T.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		_, err := unmarshalPayload[doc](json.RawMessage(`{not valid`))
		require.Error(t, err)
	})
}

func TestPayloadFromMetadata(T *testing.T) {
	T.Parallel()

	T.Run("nil metadata", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, payloadFromMetadata[doc](nil))
	})

	T.Run("non-nil metadata", func(t *testing.T) {
		t.Parallel()
		d := &doc{Kind: "k", Title: "t"}
		result := payloadFromMetadata(d)
		require.NotNil(t, result)
		assert.Equal(t, d, result)
	})
}

func TestWrapStatusError(T *testing.T) {
	T.Parallel()

	T.Run("wraps ErrUnexpectedStatus", func(t *testing.T) {
		t.Parallel()
		err := wrapStatusError(500, []byte("internal error"))
		require.ErrorIs(t, err, ErrUnexpectedStatus)
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "internal error")
	})
}

func TestCollectionPath(T *testing.T) {
	T.Parallel()

	T.Run("no suffix", func(t *testing.T) {
		t.Parallel()
		im := &indexManager[doc]{baseURL: "http://localhost:6333", collection: "my_col"}
		assert.Equal(t, "http://localhost:6333/collections/my_col", im.collectionPath(""))
	})

	T.Run("with suffix", func(t *testing.T) {
		t.Parallel()
		im := &indexManager[doc]{baseURL: "http://localhost:6333", collection: "my_col"}
		assert.Equal(t, "http://localhost:6333/collections/my_col/points?wait=true", im.collectionPath("/points?wait=true"))
	})

	T.Run("collection name is URL-escaped", func(t *testing.T) {
		t.Parallel()
		im := &indexManager[doc]{baseURL: "http://localhost:6333", collection: "has space"}
		assert.Equal(t, "http://localhost:6333/collections/has%20space", im.collectionPath(""))
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

	T.Run("invalid config missing base URL", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](t.Context(), nil, nil, nil, &Config{
			Dimension: 3,
			Metric:    vectorsearch.DistanceCosine,
		}, "test", cbnoop.NewCircuitBreaker())
		require.Error(t, err)
	})

	T.Run("collection already exists", func(t *testing.T) {
		t.Parallel()
		// GET returns 200 so ensureCollection skips creation
		stub := &qdrantStub{collectionGetStatus: http.StatusOK}
		srv := httptest.NewServer(stub)
		t.Cleanup(srv.Close)

		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
	})

	T.Run("ensureCollection GET fails with unexpected status", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{collectionGetStatus: http.StatusInternalServerError}
		srv := httptest.NewServer(stub)
		t.Cleanup(srv.Close)

		_, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.Error(t, err)
	})

	T.Run("ensureCollection PUT fails", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{
			collectionGetStatus: http.StatusNotFound,
			collectionPutStatus: http.StatusInternalServerError,
		}
		srv := httptest.NewServer(stub)
		t.Cleanup(srv.Close)

		_, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.Error(t, err)
	})

	T.Run("ensureCollection unreachable server", func(t *testing.T) {
		t.Parallel()
		_, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: "http://127.0.0.1:1", Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: 100 * time.Millisecond},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.Error(t, err)
	})

	T.Run("default timeout when zero", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{collectionGetStatus: http.StatusOK}
		srv := httptest.NewServer(stub)
		t.Cleanup(srv.Close)

		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: 0},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		require.NotNil(t, idx)
	})

	T.Run("sets api key header", func(t *testing.T) {
		t.Parallel()
		var gotAPIKey string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAPIKey = r.Header.Get("api-key")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":true,"status":"ok","time":0}`))
		}))
		t.Cleanup(srv.Close)

		_, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, APIKey: "secret", Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		assert.Equal(t, "secret", gotAPIKey)
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

func TestUpsert(T *testing.T) {
	T.Parallel()

	T.Run("empty vectors is a no-op", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		require.NoError(t, idx.Upsert(t.Context()))
	})

	T.Run("circuit breaker broken", func(t *testing.T) {
		t.Parallel()
		cb := &cbmock.CircuitBreakerMock{
			CannotProceedFunc: func() bool { return true },
		}

		idx := buildStubIndex(t, &qdrantStub{}, cb)

		err := idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}})
		require.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)

		require.Len(t, cb.CannotProceedCalls(), 1)
	})

	T.Run("rejects empty ID", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		err := idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "", Embedding: []float32{1, 0, 0}})
		require.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("rejects empty embedding", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		err := idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "a", Embedding: nil})
		require.ErrorIs(t, err, vectorsearch.ErrEmptyEmbedding)
	})

	T.Run("rejects wrong dimension", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		err := idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0}})
		require.ErrorIs(t, err, vectorsearch.ErrDimensionMismatch)
	})

	T.Run("successful upsert", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		err := idx.Upsert(t.Context(),
			vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}, Metadata: &doc{Kind: "doc", Title: "alpha"}},
			vectorsearch.Vector[doc]{ID: "b", Embedding: []float32{0, 1, 0}},
		)
		require.NoError(t, err)
	})

	T.Run("server returns error status", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{pointsPutStatus: http.StatusInternalServerError}, nil)
		err := idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnexpectedStatus)
	})

	T.Run("unreachable server", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{}
		srv := httptest.NewServer(stub)
		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		// Close the server to simulate unreachable
		srv.Close()

		err = idx.Upsert(t.Context(), vectorsearch.Vector[doc]{ID: "a", Embedding: []float32{1, 0, 0}})
		require.Error(t, err)
	})
}

func TestDelete(T *testing.T) {
	T.Parallel()

	T.Run("empty ids is a no-op", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		require.NoError(t, idx.Delete(t.Context()))
	})

	T.Run("circuit breaker broken", func(t *testing.T) {
		t.Parallel()
		cb := &cbmock.CircuitBreakerMock{
			CannotProceedFunc: func() bool { return true },
		}

		idx := buildStubIndex(t, &qdrantStub{}, cb)

		err := idx.Delete(t.Context(), "some-id")
		require.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)

		require.Len(t, cb.CannotProceedCalls(), 1)
	})

	T.Run("successful delete", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		require.NoError(t, idx.Delete(t.Context(), "id1", "id2"))
	})

	T.Run("server returns error status", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{pointsDeleteStatus: http.StatusInternalServerError}, nil)
		err := idx.Delete(t.Context(), "id1")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnexpectedStatus)
	})

	T.Run("unreachable server", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{}
		srv := httptest.NewServer(stub)
		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		srv.Close()

		err = idx.Delete(t.Context(), "id1")
		require.Error(t, err)
	})
}

func TestWipe(T *testing.T) {
	T.Parallel()

	T.Run("circuit breaker broken", func(t *testing.T) {
		t.Parallel()
		cb := &cbmock.CircuitBreakerMock{
			CannotProceedFunc: func() bool { return true },
		}

		idx := buildStubIndex(t, &qdrantStub{}, cb)

		err := idx.Wipe(t.Context())
		require.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)

		require.Len(t, cb.CannotProceedCalls(), 1)
	})

	T.Run("successful wipe", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		require.NoError(t, idx.Wipe(t.Context()))
	})

	T.Run("delete collection fails", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{collectionDeleteStatus: http.StatusForbidden}, nil)
		err := idx.Wipe(t.Context())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnexpectedStatus)
	})

	T.Run("delete returns 404 still succeeds", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{collectionDeleteStatus: http.StatusNotFound}, nil)
		require.NoError(t, idx.Wipe(t.Context()))
	})

	T.Run("recreate collection fails after delete", func(t *testing.T) {
		t.Parallel()
		callCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			// Initial GET during ProvideIndex — collection exists
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/collections/"):
				if callCount == 0 {
					// First GET (ProvideIndex ensureCollection) — exists
					callCount++
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"result":true}`))
				} else {
					// Second GET (Wipe recreate ensureCollection) — not found so it tries to PUT
					w.WriteHeader(http.StatusNotFound)
				}
			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/collections/"):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":true}`))
			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/collections/"):
				// Recreate fails
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"error"}`))
			default:
				http.Error(w, "unexpected", http.StatusBadRequest)
			}
		}))
		t.Cleanup(srv.Close)

		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)

		err = idx.Wipe(t.Context())
		require.Error(t, err)
	})

	T.Run("unreachable server", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{}
		srv := httptest.NewServer(stub)
		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		srv.Close()

		err = idx.Wipe(t.Context())
		require.Error(t, err)
	})
}

func TestQuery(T *testing.T) {
	T.Parallel()

	T.Run("rejects empty embedding", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: nil, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrEmptyEmbedding)
	})

	T.Run("rejects wrong dimension", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0}, TopK: 5})
		require.ErrorIs(t, err, vectorsearch.ErrDimensionMismatch)
	})

	T.Run("circuit breaker broken", func(t *testing.T) {
		t.Parallel()
		cb := &cbmock.CircuitBreakerMock{
			CannotProceedFunc: func() bool { return true },
		}

		idx := buildStubIndex(t, &qdrantStub{}, cb)

		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)

		require.Len(t, cb.CannotProceedCalls(), 1)
	})

	T.Run("defaults TopK to 10", func(t *testing.T) {
		t.Parallel()
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/points/search") {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":[]}`))
				return
			}
			// ensureCollection
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":true}`))
		}))
		t.Cleanup(srv.Close)

		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)

		_, err = idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 0})
		require.NoError(t, err)
		require.NotNil(t, gotBody)
		assert.Equal(t, float64(10), gotBody["limit"])
	})

	T.Run("successful query returns results", func(t *testing.T) {
		t.Parallel()
		searchResp := `{"result":[{"id":"abc","score":0.95,"payload":{"kind":"doc","title":"hello"}},{"id":"def","score":0.8,"payload":null}]}`
		idx := buildStubIndex(t, &qdrantStub{pointsSearchBody: searchResp}, nil)

		results, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.NoError(t, err)
		require.Len(t, results, 2)

		assert.Equal(t, "abc", results[0].ID)
		assert.InDelta(t, 0.95, float64(results[0].Distance), 0.001)
		require.NotNil(t, results[0].Metadata)
		assert.Equal(t, "doc", results[0].Metadata.Kind)
		assert.Equal(t, "hello", results[0].Metadata.Title)

		assert.Equal(t, "def", results[1].ID)
		assert.Nil(t, results[1].Metadata)
	})

	T.Run("query with numeric ID", func(t *testing.T) {
		t.Parallel()
		searchResp := `{"result":[{"id":42,"score":0.5,"payload":null}]}`
		idx := buildStubIndex(t, &qdrantStub{pointsSearchBody: searchResp}, nil)

		results, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 1})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "42", results[0].ID)
	})

	T.Run("query with filter", func(t *testing.T) {
		t.Parallel()
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/points/search") {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"result":[]}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":true}`))
		}))
		t.Cleanup(srv.Close)

		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)

		filter := map[string]any{"must": []any{map[string]any{"key": "kind", "match": map[string]any{"value": "doc"}}}}
		_, err = idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5, Filter: filter})
		require.NoError(t, err)
		require.NotNil(t, gotBody)
		assert.NotNil(t, gotBody["filter"])
	})

	T.Run("server returns error status", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{pointsSearchStatus: http.StatusInternalServerError}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnexpectedStatus)
	})

	T.Run("invalid JSON response", func(t *testing.T) {
		t.Parallel()
		idx := buildStubIndex(t, &qdrantStub{pointsSearchBody: `{not json`}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.Error(t, err)
	})

	T.Run("invalid payload in response", func(t *testing.T) {
		t.Parallel()
		// payload is a string where a doc struct is expected
		searchResp := `{"result":[{"id":"x","score":0.5,"payload":"not-a-doc"}]}`
		idx := buildStubIndex(t, &qdrantStub{pointsSearchBody: searchResp}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.Error(t, err)
	})

	T.Run("unsupported ID type in response", func(t *testing.T) {
		t.Parallel()
		searchResp := `{"result":[{"id":true,"score":0.5,"payload":null}]}`
		idx := buildStubIndex(t, &qdrantStub{pointsSearchBody: searchResp}, nil)
		_, err := idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.Error(t, err)
	})

	T.Run("unreachable server", func(t *testing.T) {
		t.Parallel()
		stub := &qdrantStub{}
		srv := httptest.NewServer(stub)
		idx, err := ProvideIndex[doc](
			t.Context(), nil, nil, nil,
			&Config{BaseURL: srv.URL, Dimension: 3, Metric: vectorsearch.DistanceCosine, Timeout: time.Second},
			"test",
			cbnoop.NewCircuitBreaker(),
		)
		require.NoError(t, err)
		srv.Close()

		_, err = idx.Query(t.Context(), vectorsearch.QueryRequest{Embedding: []float32{1, 0, 0}, TopK: 5})
		require.Error(t, err)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		var cfg *Config
		err := cfg.ValidateWithContext(t.Context())
		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BaseURL:   "http://localhost:6333",
			Dimension: 128,
			Metric:    vectorsearch.DistanceCosine,
		}
		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing base URL", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Dimension: 128,
			Metric:    vectorsearch.DistanceCosine,
		}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing dimension", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BaseURL: "http://localhost:6333",
			Metric:  vectorsearch.DistanceCosine,
		}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("missing metric", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BaseURL:   "http://localhost:6333",
			Dimension: 128,
		}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("invalid metric value", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BaseURL:   "http://localhost:6333",
			Dimension: 128,
			Metric:    "invalid",
		}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("all valid metrics pass", func(t *testing.T) {
		t.Parallel()
		for _, m := range []vectorsearch.DistanceMetric{
			vectorsearch.DistanceCosine,
			vectorsearch.DistanceDotProduct,
			vectorsearch.DistanceEuclidean,
		} {
			cfg := &Config{
				BaseURL:   "http://localhost:6333",
				Dimension: 128,
				Metric:    m,
			}
			require.NoError(t, cfg.ValidateWithContext(t.Context()), "metric %q should be valid", m)
		}
	})
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
