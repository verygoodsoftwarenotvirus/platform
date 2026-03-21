package mock

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/llm"

	"github.com/stretchr/testify/require"
)

func TestProvider_Completion(t *testing.T) {
	t.Parallel()

	m := &Provider{}
	m.On("Completion", t.Context(), llm.CompletionParams{
		Model:    "test",
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	}).Return(&llm.CompletionResult{Content: "mocked"}, nil)

	ctx := t.Context()
	result, err := m.Completion(ctx, llm.CompletionParams{
		Model:    "test",
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})

	require.NoError(t, err)
	require.Equal(t, "mocked", result.Content)
	m.AssertExpectations(t)
}
