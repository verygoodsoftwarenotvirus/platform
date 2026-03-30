package kubectl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("valid config with kubeconfig", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "production", Kubeconfig: "/home/user/.kube/config"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("missing namespace", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		err := cfg.ValidateWithContext(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "namespace")
	})
}
