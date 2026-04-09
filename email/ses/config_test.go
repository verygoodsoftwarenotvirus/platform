package ses

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Region: "us-east-1",
		}

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with missing region", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		err := cfg.ValidateWithContext(t.Context())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})
}
