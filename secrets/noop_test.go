package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source := NewNoopSecretSource()
		ctx := context.Background()

		got, err := source.GetSecret(ctx, "any-key")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestNoopSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source := NewNoopSecretSource()
		err := source.Close()
		require.NoError(t, err)
	})
}
