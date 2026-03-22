package posthog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/posthog/posthog-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFeatureFlagManager(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing API key", func(t *testing.T) {
		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing API key", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid config", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Interval = -1
		})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
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

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, flagName)
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error executing request", func(t *testing.T) {
		ctx := t.Context()
		exampleUsername := "username"

		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusForbidden)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, t.Name())
		assert.Error(t, err)
		assert.False(t, actual)
	})
}
