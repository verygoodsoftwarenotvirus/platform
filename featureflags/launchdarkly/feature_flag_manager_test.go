package launchdarkly

import (
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	mockCircuitBreaker "github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	ld "github.com/launchdarkly/go-server-sdk/v6"
	"github.com/launchdarkly/go-server-sdk/v6/subsystems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeLaunchDarklyDataSource struct{}

func (f *fakeLaunchDarklyDataSource) Close() error {
	return nil
}

func (f *fakeLaunchDarklyDataSource) IsInitialized() bool {
	return true
}

func (f *fakeLaunchDarklyDataSource) Start(closeWhenReady chan<- struct{}) {
	close(closeWhenReady)
}

type fakeLaunchDarklyDataSourceBuilder struct{}

// Build is called internally by the SDK.
func (b *fakeLaunchDarklyDataSourceBuilder) Build(subsystems.ClientContext) (subsystems.DataSource, error) {
	return &fakeLaunchDarklyDataSource{}, nil
}

func buildTestManager(t *testing.T, cb circuitbreaking.CircuitBreaker) *featureFlagManager {
	t.Helper()

	cfg := &Config{SDKKey: t.Name()}

	ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, cb, func(config ld.Config) ld.Config {
		config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
		return config
	})
	require.NoError(t, err)
	require.NotNil(t, ffm)

	return ffm.(*featureFlagManager)
}

func TestNewFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{SDKKey: t.Name()}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("with missing http launchDarklyClient", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{SDKKey: t.Name()}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with missing SDK key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.Error(t, err)
		require.Nil(t, actual)
	})

	T.Run("with zero init timeout gets default", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{SDKKey: t.Name(), InitTimeout: 0}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker(), func(config ld.Config) ld.Config {
			config.DataSource = &fakeLaunchDarklyDataSourceBuilder{}
			return config
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(true)
		cb.On("Succeeded").Return()
		cb.On("Failed").Return()

		ffm := buildTestManager(t, cb)

		result, _ := ffm.CanUseFeature(ctx, "user123", "some-flag")
		assert.False(t, result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(false)

		ffm := buildTestManager(t, cb)

		result, err := ffm.CanUseFeature(ctx, "user123", "some-flag")
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.False(t, result)
	})
}

func TestFeatureFlagManager_GetStringValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(true)
		cb.On("Succeeded").Return()
		cb.On("Failed").Return()

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetStringValue(ctx, "user123", "some-flag")
		_ = err
		assert.Empty(t, result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(false)

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetStringValue(ctx, "user123", "some-flag")
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Empty(t, result)
	})
}

func TestFeatureFlagManager_GetInt64Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(true)
		cb.On("Succeeded").Return()
		cb.On("Failed").Return()

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetInt64Value(ctx, "user123", "some-flag")
		_ = err
		assert.Zero(t, result)
	})

	T.Run("with broken circuit", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cb := &mockCircuitBreaker.MockCircuitBreaker{}
		cb.On("CanProceed").Return(false)

		ffm := buildTestManager(t, cb)

		result, err := ffm.GetInt64Value(ctx, "user123", "some-flag")
		assert.ErrorIs(t, err, circuitbreaking.ErrCircuitBroken)
		assert.Zero(t, result)
	})
}

func TestFeatureFlagManager_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ffm := buildTestManager(t, circuitbreaking.NewNoopCircuitBreaker())

		err := ffm.Close()
		assert.NoError(t, err)
	})
}
