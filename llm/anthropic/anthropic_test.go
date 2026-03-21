package anthropic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/llm"

	"github.com/stretchr/testify/require"
)

// anthropicMessageResponse mimics the Anthropic Messages API response format.
func anthropicMessageResponse(content string) map[string]any {
	return map[string]any{
		"id":          "msg-test",
		"type":        "message",
		"role":        "assistant",
		"model":       "claude-sonnet-4-20250514",
		"content":     []map[string]any{{"type": "text", "text": content}},
		"stop_reason": "end_turn",
		"usage": map[string]any{
			"input_tokens":  10,
			"output_tokens": 5,
		},
	}
}

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

	T.Run("with base URL", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(&Config{
			APIKey:       "test-key",
			BaseURL:      "https://custom.example.com",
			DefaultModel: "claude-sonnet-4",
		})
		require.NoError(t, err)
		require.NotNil(t, provider)
	})
}

func TestAnthropicProvider_Completion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/v1/messages", r.URL.Path)
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(anthropicMessageResponse("Hello from Claude mock!")))
		}))
		t.Cleanup(ts.Close)

		provider, err := NewProvider(&Config{
			APIKey:  "test-key",
			BaseURL: ts.URL,
		})
		require.NoError(t, err)
		require.NotNil(t, provider)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Model: "claude-sonnet-4-20250514",
			Messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "Hello from Claude mock!", result.Content)
	})

	T.Run("uses default model when not specified", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(anthropicMessageResponse("Hi there!")))
		}))
		t.Cleanup(ts.Close)

		provider, err := NewProvider(&Config{
			APIKey:       "test-key",
			BaseURL:      ts.URL,
			DefaultModel: "claude-sonnet-4",
		})
		require.NoError(t, err)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Messages: []llm.Message{{Role: "user", Content: "Hi"}},
		})
		require.NoError(t, err)
		require.Equal(t, "Hi there!", result.Content)
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
			BaseURL: ts.URL,
		})
		require.NoError(t, err)

		ctx := t.Context()
		result, err := provider.Completion(ctx, llm.CompletionParams{
			Model:    "claude-sonnet-4-20250514",
			Messages: []llm.Message{{Role: "user", Content: "Hi"}},
		})
		require.Error(t, err)
		require.Nil(t, result)
	})
}
