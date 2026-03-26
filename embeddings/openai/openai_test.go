package openai

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

	T.Run("with missing API key", func(t *testing.T) {
		t.Parallel()

		emb, err := NewEmbedder(t.Context(), &Config{}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.Error(t, err)
		require.Nil(t, emb)
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		emb, err := NewEmbedder(t.Context(), &Config{APIKey: "test-key"}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)
		require.NotNil(t, emb)
	})
}

func TestEmbedder_GenerateEmbedding(T *testing.T) {
	T.Parallel()

	openAIEmbeddingResponse := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{
				"object":    "embedding",
				"index":     0,
				"embedding": []float64{0.1, 0.2, 0.3},
			},
		},
		"model": "text-embedding-3-small",
		"usage": map[string]any{
			"prompt_tokens": 5,
			"total_tokens":  5,
		},
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1/embeddings", r.URL.Path)
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(openAIEmbeddingResponse))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			APIKey:  "test-key",
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
		assert.Equal(t, "text-embedding-3-small", result.Model)
		assert.Equal(t, "openai", result.Provider)
		assert.Equal(t, 3, result.Dimensions)
		assert.Len(t, result.Vector, 3)
		assert.False(t, result.GeneratedAt.IsZero())
	})

	T.Run("uses input model override", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody embeddingRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
			require.Equal(t, "text-embedding-3-large", reqBody.Model)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(openAIEmbeddingResponse))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			APIKey:       "test-key",
			BaseURL:      ts.URL,
			DefaultModel: "text-embedding-3-small",
		}, logging.NewNoopLogger(), tracing.NewTracerForTest("test"))
		require.NoError(t, err)

		ctx := t.Context()
		result, err := emb.GenerateEmbedding(ctx, &embeddings.Input{
			Content: "hello",
			Model:   "text-embedding-3-large",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	T.Run("with non-200 response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		t.Cleanup(ts.Close)

		emb, err := NewEmbedder(t.Context(), &Config{
			APIKey:  "test-key",
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
			APIKey:  "test-key",
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
