package posthog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{APIKey: t.Name()}

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with empty API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}
