package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_BuildClient(T *testing.T) {
	T.Parallel()

	T.Run("with tracing enabled", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Timeout:       2 * time.Second,
			EnableTracing: true,
		}
		cfg.EnsureDefaults()

		client := cfg.BuildClient()
		require.NotNil(t, client)
		assert.Equal(t, 2*time.Second, client.Timeout)
		assert.NotNil(t, client.Transport)
	})

	T.Run("with tracing disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Timeout:       3 * time.Second,
			EnableTracing: false,
		}
		cfg.EnsureDefaults()

		client := cfg.BuildClient()
		require.NotNil(t, client)
		assert.Equal(t, 3*time.Second, client.Timeout)
		assert.NotNil(t, client.Transport)
	})

	T.Run("applies MaxIdleConns and MaxIdleConnsPerHost", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Timeout:             time.Second,
			MaxIdleConns:        42,
			MaxIdleConnsPerHost: 21,
			EnableTracing:       false,
		}
		cfg.EnsureDefaults()

		client := cfg.BuildClient()
		require.NotNil(t, client)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		assert.Equal(t, 42, transport.MaxIdleConns)
		assert.Equal(t, 21, transport.MaxIdleConnsPerHost)
	})
}

func TestProvideHTTPClient(T *testing.T) {
	T.Parallel()

	T.Run("with nil config uses defaults", func(t *testing.T) {
		t.Parallel()

		client := ProvideHTTPClient(nil)
		require.NotNil(t, client)
		assert.Equal(t, defaultTimeout, client.Timeout)
	})

	T.Run("with config uses config values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Timeout: 7 * time.Second,
		}
		client := ProvideHTTPClient(cfg)
		require.NotNil(t, client)
		assert.Equal(t, 7*time.Second, client.Timeout)
	})
}
