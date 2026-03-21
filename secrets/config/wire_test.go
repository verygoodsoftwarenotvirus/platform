package secretscfg

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvideSecretSourceFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("nil config returns env source", func(t *testing.T) {
		t.Parallel()

		var cfg *Config
		source, err := ProvideSecretSourceFromConfig(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_WIRE_NIL_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		require.Equal(t, value, got)
	})

	t.Run("empty provider returns env source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ""}
		source, err := ProvideSecretSourceFromConfig(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_WIRE_EMPTY_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		require.Equal(t, value, got)
	})

	t.Run("noop provider returns noop source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderNoop}
		source, err := ProvideSecretSourceFromConfig(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, source)

		got, err := source.GetSecret(context.Background(), "any")
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("env provider returns env source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderEnv}
		source, err := ProvideSecretSourceFromConfig(context.Background(), cfg)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_WIRE_ENV_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		require.Equal(t, value, got)
	})
}
