package asyncnotificationscfg

import (
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications/ably"
	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications/pusher"
	asyncws "github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications/websocket"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderNoop,
		}

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "invalid",
		}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("pusher requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderPusher,
		}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("ably requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderAbly,
		}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("websocket requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderWebSocket,
		}

		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestConfig_ProvideAsyncNotifier(T *testing.T) {
	T.Parallel()

	T.Run("with websocket", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:  ProviderWebSocket,
			WebSocket: &asyncws.Config{},
		}

		actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with sse", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderSSE,
		}

		actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with pusher", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderPusher,
			Pusher: &pusher.Config{
				AppID:   "123",
				Key:     "key",
				Secret:  "secret",
				Cluster: "us2",
			},
		}

		actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with ably", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderAbly,
			Ably: &ably.Config{
				APIKey: "appid.keyid:keysecret",
			},
		}

		actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	noopProviders := []string{"", ProviderNoop}
	for _, provider := range noopProviders {
		T.Run(fmt.Sprintf("with noop provider %q", provider), func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				Provider: provider,
			}

			actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
			assert.NotNil(t, actual)
			assert.NoError(t, err)
		})
	}

	T.Run("with unknown provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "unknown",
		}

		actual, err := cfg.ProvideAsyncNotifier(logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
