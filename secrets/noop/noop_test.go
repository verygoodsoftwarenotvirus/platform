package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source := NewSecretSource()
		ctx := context.Background()

		got, err := source.GetSecret(ctx, "any-key")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source := NewSecretSource()
		err := source.Close()
		require.NoError(t, err)
	})
}
