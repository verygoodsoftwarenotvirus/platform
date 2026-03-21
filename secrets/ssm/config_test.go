package ssm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	t.Run("invalid missing Region", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: ""}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})
}
