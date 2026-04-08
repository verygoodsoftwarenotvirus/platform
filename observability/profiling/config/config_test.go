package profilingcfg

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/profiling/pprof"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/profiling/pyroscope"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid empty provider", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ""}
		assert.NoError(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("valid pprof provider", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPprof,
			Pprof:    &pprof.Config{Port: 6060},
		}
		assert.NoError(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("valid pyroscope provider", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPyroscope,
			Pyroscope: &pyroscope.Config{
				ServerAddress: "http://localhost:4040",
				UploadRate:    1,
			},
		}
		assert.NoError(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("invalid provider string", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: "invalid"}
		assert.Error(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("pyroscope provider without config", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPyroscope}
		assert.Error(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("pprof config present with empty provider is invalid", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: "",
			Pprof:    &pprof.Config{Port: 6060},
		}
		assert.Error(t, c.ValidateWithContext(t.Context()))
	})

	T.Run("pyroscope config present with pprof provider is invalid", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPprof,
			Pyroscope: &pyroscope.Config{
				ServerAddress: "http://localhost:4040",
				UploadRate:    1,
			},
		}
		assert.Error(t, c.ValidateWithContext(t.Context()))
	})
}

func TestConfig_ProvideProfilingProvider(T *testing.T) {
	T.Parallel()

	logger := logging.NewNoopLogger()

	T.Run("default provider returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ""}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("unknown provider returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: "unknown"}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("pprof with nil config uses defaults", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPprof}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(t.Context()))
	})

	T.Run("pprof with config", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPprof,
			Pprof:    &pprof.Config{Port: 16060},
		}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(t.Context()))
	})

	T.Run("pyroscope with nil config returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPyroscope}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("pyroscope with config sets default upload rate", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider:    ProviderPyroscope,
			ServiceName: "test-service",
			Pyroscope: &pyroscope.Config{
				ServerAddress: "http://localhost:4040",
			},
		}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, 15*time.Second, c.Pyroscope.UploadRate)
		require.NoError(t, p.Shutdown(t.Context()))
	})

	T.Run("pyroscope with non-default upload rate", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider:    ProviderPyroscope,
			ServiceName: "test-service",
			Pyroscope: &pyroscope.Config{
				ServerAddress: "http://localhost:4040",
				UploadRate:    5 * time.Second,
			},
		}
		p, err := c.ProvideProfilingProvider(t.Context(), logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, 5*time.Second, c.Pyroscope.UploadRate)
		require.NoError(t, p.Shutdown(t.Context()))
	})
}
