package pprof

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideProfilingProvider(T *testing.T) {
	T.Parallel()

	T.Run("nil config returns noop", func(t *testing.T) {
		t.Parallel()
		p, err := ProvideProfilingProvider(context.Background(), logging.NewNoopLogger(), nil)
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	T.Run("zero port uses default", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Port: 0}
		p, err := ProvideProfilingProvider(context.Background(), logging.NewNoopLogger(), cfg)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(context.Background()))
	})

	T.Run("with mutex and block profiling", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			Port:               16061,
			EnableMutexProfile: true,
			EnableBlockProfile: true,
		}
		p, err := ProvideProfilingProvider(context.Background(), logging.NewNoopLogger(), cfg)
		require.NoError(t, err)
		assert.NotNil(t, p)
		require.NoError(t, p.Shutdown(context.Background()))
	})

	T.Run("start and shutdown", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Port: 16062}
		p, err := ProvideProfilingProvider(context.Background(), logging.NewNoopLogger(), cfg)
		require.NoError(t, err)
		require.NoError(t, p.Start(context.Background()))
		require.NoError(t, p.Shutdown(context.Background()))
	})
}
