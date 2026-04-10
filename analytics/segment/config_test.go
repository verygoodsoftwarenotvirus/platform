package segment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{APIToken: t.Name()}

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with empty API token", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}
