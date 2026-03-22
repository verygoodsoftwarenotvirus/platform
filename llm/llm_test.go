package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopProvider_Completion(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ctx := t.Context()
		provider := NewNoopProvider()

		result, err := provider.Completion(ctx, CompletionParams{
			Model: "test",
			Messages: []Message{
				{Role: "user", Content: "hello"},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
