package profilingcfg

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/profiling/pprof"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/profiling/pyroscope"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	ctx := context.Background()

	T.Run("valid empty provider", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ""}
		assert.NoError(t, c.ValidateWithContext(ctx))
	})

	T.Run("valid pprof provider", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPprof,
			Pprof:    &pprof.Config{Port: 6060},
		}
		assert.NoError(t, c.ValidateWithContext(ctx))
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
		assert.NoError(t, c.ValidateWithContext(ctx))
	})

	T.Run("invalid provider string", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: "invalid"}
		assert.Error(t, c.ValidateWithContext(ctx))
	})

	T.Run("pyroscope provider without config", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPyroscope}
		assert.Error(t, c.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideProfilingProvider(T *testing.T) {
	T.Parallel()

	ctx := context.Background()

	logger := logging.NewNoopLogger()

	T.Run("default provider returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ""}
		p, err := c.ProvideProfilingProvider(ctx, logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("unknown provider returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: "unknown"}
		p, err := c.ProvideProfilingProvider(ctx, logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("pprof with nil config uses defaults", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPprof}
		p, err := c.ProvideProfilingProvider(ctx, logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(ctx))
	})

	T.Run("pprof with config", func(t *testing.T) {
		t.Parallel()
		c := &Config{
			Provider: ProviderPprof,
			Pprof:    &pprof.Config{Port: 16060},
		}
		p, err := c.ProvideProfilingProvider(ctx, logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(ctx))
	})

	T.Run("pyroscope with nil config returns noop", func(t *testing.T) {
		t.Parallel()
		c := &Config{Provider: ProviderPyroscope}
		p, err := c.ProvideProfilingProvider(ctx, logger)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}
