package redis

import (
	"testing"

	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()
		var cfg *Config
		require.ErrorIs(t, cfg.ValidateWithContext(t.Context()), platformerrors.ErrNilInputParameter)
	})

	T.Run("missing addresses", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		err := cfg.ValidateWithContext(t.Context())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "addresses")
	})

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Addresses: []string{"localhost:6379"}}
		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})
}
