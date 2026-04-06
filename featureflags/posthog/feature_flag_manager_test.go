package posthog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/circuitbreaking"
	mockCircuitBreaker "github.com/verygoodsoftwarenotvirus/platform/v4/circuitbreaking/mock"
	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v4/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/posthog/posthog-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestManager(t *testing.T, cb circuitbreaking.CircuitBreaker, configModifiers ...func(config *posthog.Config)) *featureFlagManager {
	t.Helper()

	cfg := &Config{
		ProjectAPIKey:  t.Name(),
		PersonalAPIKey: t.Name(),
	}

	ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cb, configModifiers...)
	require.NoError(t, err)
	require.NotNil(t, ffm)

	return ffm.(*featureFlagManager)
}

func TestNewFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing project API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing personal API key", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker(), func(config *posthog.Config) {
			config.Interval = -1
		})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		t.SkipNow()

		ctx := t.Context()
		exampleUsername := "username"

		flagName := t.Name()
		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			response, err := json.Marshal(&posthog.FeatureFlagsResponse{
				Flags: []posthog.FeatureFlag{
					{
						Key:    flagName,
						Active: true,
						Filters: posthog.Filter{
							Groups: []posthog.FeatureFlagCondition{
								{
									Properties:        nil,
									RolloutPercentage: nil,
									Variant:           nil,
								},
							},
						},
					},
				},
				GroupTypeMapping: new(map[string]string{}),
			})
			require.NoError(t, err)

			_, err = res.Write(response)
			require.NoError(t, err)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, flagName)
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUsername := "username"

		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusForbidden)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cbnoop.NewCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, t.Name())
		assert.Error(t, err)
		assert.False(t, actual)
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

		ffm := buildTestManager(t, cbnoop.NewCircuitBreaker())

		err := ffm.Close()
		assert.NoError(t, err)
	})
}
