package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/llm"

	"github.com/stretchr/testify/require"
)

func TestNewProvider(T *testing.T) {
	T.Parallel()

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(nil)
		require.Error(t, err)
		require.Nil(t, provider)
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(&Config{APIKey: "test-key"})
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	T.Run("with base URL and timeout", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(&Config{
			APIKey:       "test-key",
			BaseURL:      "https://custom.example.com/v1",
			DefaultModel: "gpt-4o",
		})
		require.NoError(t, err)
		require.NotNil(t, provider)
	})
}

func TestOpenAIProvider_Completion(T *testing.T) {
	T.Parallel()

	openAIChatCompletion := map[string]any{
		"id":      "chatcmpl-test",
		"object":  "chat.completion",
		"created": 1234567890,
		"model":   "gpt-4o-mini",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": "Hello from mock!",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     10,
			"completion_tokens": 5,
			"total_tokens":      15,
		},
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1/chat/completions", r.URL.Path)
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(openAIChatCompletion))
		}))
		t.Cleanup(ts.Close)

		provider, err := NewProvider(&Config{
			APIKey:  "test-key",
			BaseURL: ts.URL + "/v1",
		})
		require.NoError(t, err)
		require.NotNil(t, provider)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Model: "gpt-4o-mini",
			Messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Hello from mock!", result.Content)
	})

	T.Run("uses default model when not specified", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(openAIChatCompletion))
		}))
		t.Cleanup(ts.Close)

		provider, err := NewProvider(&Config{
			APIKey:       "test-key",
			BaseURL:      ts.URL + "/v1",
			DefaultModel: "gpt-4o",
		})
		require.NoError(t, err)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Messages: []llm.Message{{Role: "user", Content: "Hi"}},
		})
		require.NoError(t, err)
		require.Equal(t, "Hello from mock!", result.Content)
	})

	T.Run("with API error", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		t.Cleanup(ts.Close)

		provider, err := NewProvider(&Config{
			APIKey:  "test-key",
			BaseURL: ts.URL + "/v1",
		})
		require.NoError(t, err)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Model:    "gpt-4o-mini",
			Messages: []llm.Message{{Role: "user", Content: "Hi"}},
		})
		require.Error(t, err)
		require.Nil(t, result)
	})
}
