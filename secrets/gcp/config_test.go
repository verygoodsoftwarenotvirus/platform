package gcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "my-project"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	t.Run("invalid missing ProjectID", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: ""}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})
}
