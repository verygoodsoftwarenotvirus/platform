package env

import (
	"context"
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/secrets"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ secrets.SecretSource = (*envSecretSource)(nil)

func TestEnvSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("returns set env var", func(t *testing.T) {
		t.Parallel()

		key := "TEST_SECRET_" + t.Name()
		value := "secret-value"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		source := NewEnvSecretSource()
		ctx := context.Background()

		got, err := source.GetSecret(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	T.Run("returns empty for unset env var", func(t *testing.T) {
		t.Parallel()

		key := "TEST_SECRET_UNSET_" + t.Name()
		require.NoError(t, os.Unsetenv(key))

		source := NewEnvSecretSource()
		ctx := context.Background()

		got, err := source.GetSecret(ctx, key)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestEnvSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source := NewEnvSecretSource()
		err := source.Close()
		require.NoError(t, err)
	})
}
