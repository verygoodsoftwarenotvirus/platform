package gcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "my-project"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("invalid missing ProjectID", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: ""}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})
}
