package ssm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("invalid missing Region", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: ""}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})
}
