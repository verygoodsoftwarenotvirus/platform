package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("SSE provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderSSE,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("WebSocket provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderWebSocket,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx), "websocket provider requires websocket config")
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "invalid",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideEventStreamUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("SSE", func(t *testing.T) {
		t.Parallel()

		upgrader, err := ProvideEventStreamUpgrader(&Config{
			Provider: ProviderSSE,
		})

		require.NoError(t, err)
		assert.NotNil(t, upgrader)
	})

	T.Run("WebSocket", func(t *testing.T) {
		t.Parallel()

		upgrader, err := ProvideEventStreamUpgrader(&Config{
			Provider: ProviderWebSocket,
		})

		require.NoError(t, err)
		assert.NotNil(t, upgrader)
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideEventStreamUpgrader(&Config{})

		assert.Error(t, err)
	})
}

func TestProvideBidirectionalEventStreamUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("SSE returns error", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideBidirectionalEventStreamUpgrader(&Config{
			Provider: ProviderSSE,
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SSE does not support bidirectional")
	})

	T.Run("WebSocket", func(t *testing.T) {
		t.Parallel()

		upgrader, err := ProvideBidirectionalEventStreamUpgrader(&Config{
			Provider: ProviderWebSocket,
		})

		require.NoError(t, err)
		assert.NotNil(t, upgrader)
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		_, err := ProvideBidirectionalEventStreamUpgrader(&Config{})

		assert.Error(t, err)
	})
}
