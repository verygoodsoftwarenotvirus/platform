package featureflagscfg

import (
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/featureflags/launchdarkly"
	"github.com/verygoodsoftwarenotvirus/platform/v2/featureflags/posthog"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			LaunchDarkly: &launchdarkly.Config{
				SDKKey:      t.Name(),
				InitTimeout: 123,
			},
			Provider: ProviderLaunchDarkly,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty provider for noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "invalid_provider",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with posthog provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			PostHog: &posthog.Config{
				ProjectAPIKey:  t.Name(),
				PersonalAPIKey: t.Name(),
			},
			Provider: ProviderPostHog,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with launchdarkly provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderLaunchDarkly,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with posthog provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderPostHog,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()
	})
}

func TestConfig_ProvideFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("with default/noop provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "",
		}

		ffm, err := cfg.ProvideFeatureFlagManager(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, ffm)
	})

	T.Run("with unknown provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "something_unknown",
		}

		ffm, err := cfg.ProvideFeatureFlagManager(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)
		require.NotNil(t, ffm)
	})

	T.Run("with launchdarkly provider but nil config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderLaunchDarkly,
		}

		ffm, err := cfg.ProvideFeatureFlagManager(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, ffm)
	})

	T.Run("with posthog provider but nil config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderPostHog,
		}

		ffm, err := cfg.ProvideFeatureFlagManager(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, ffm)
	})

	T.Run("with provider string that has whitespace and mixed case", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "  LAUNCHDARKLY  ",
		}

		// Will fail because LaunchDarkly config is nil, but proves the normalization works
		ffm, err := cfg.ProvideFeatureFlagManager(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), http.DefaultClient, circuitbreaking.NewNoopCircuitBreaker())
		require.Error(t, err)
		require.Nil(t, ffm)
	})
}

func TestProvideFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("with noop provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
		}

		ffm, err := ProvideFeatureFlagManager(ctx, cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), http.DefaultClient)
		require.NoError(t, err)
		require.NotNil(t, ffm)
	})
}
