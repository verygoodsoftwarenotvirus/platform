package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/embeddings"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmbedder(T *testing.T) {
	T.Parallel()

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		emb, err := NewEmbedder(t.Context(), nil, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.Error(t, err)
		require.Nil(t, emb)
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		emb, err := NewEmbedder(t.Context(), &Config{}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)
		require.NotNil(t, emb)
	})

	T.Run("with custom base URL", func(t *testing.T) {
		t.Parallel()

		emb, err := NewEmbedder(t.Context(), &Config{
			BaseURL: "http://custom:11434",
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)
		require.NotNil(t, emb)
	})
}

func TestEmbedder_GenerateEmbedding(T *testing.T) {
	T.Parallel()

	ollamaEmbeddingResponse := map[string]any{
		"embeddings": [][]float64{
			{0.1, 0.2, 0.3, 0.4},
		},
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/api/embed", r.URL.Path)
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(ollamaEmbeddingResponse))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			BaseURL: ts.URL,
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)

		ctx := t.Context()
		result, err := emb.GenerateEmbedding(ctx, &embeddings.Input{
			Content: "hello world",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "hello world", result.SourceText)
		assert.Equal(t, "nomic-embed-text", result.Model)
		assert.Equal(t, "ollama", result.Provider)
		assert.Equal(t, 4, result.Dimensions)
		assert.Len(t, result.Vector, 4)
		assert.False(t, result.GeneratedAt.IsZero())
	})

	T.Run("uses input model override", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody embeddingRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
			require.Equal(t, "mxbai-embed-large", reqBody.Model)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(ollamaEmbeddingResponse))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			BaseURL:      ts.URL,
			DefaultModel: "nomic-embed-text",
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)

		ctx := t.Context()
		result, err := emb.GenerateEmbedding(ctx, &embeddings.Input{
			Content: "hello",
			Model:   "mxbai-embed-large",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	T.Run("with non-200 response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"server error"}`))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			BaseURL: ts.URL,
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)

		ctx := t.Context()
		result, err := emb.GenerateEmbedding(ctx, &embeddings.Input{
			Content: "hello",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})

	T.Run("with malformed JSON response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{not json`))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			BaseURL: ts.URL,
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)

		ctx := t.Context()
		result, err := emb.GenerateEmbedding(ctx, &embeddings.Input{
			Content: "hello",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})
}
