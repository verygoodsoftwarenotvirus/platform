package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("happy path zero namespace", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("happy path explicit namespace", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: 42}
		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})
}
