package analyticscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/posthog"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/rudderstack"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/segment"
	circuitbreakingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderSegment,
				Segment:  &segment.Config{APIToken: t.Name()},
			},
		}

		require.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderSegment,
			},
		}

		require.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideCollector(T *testing.T) {
	T.Parallel()

	allProviders := []string{
		ProviderSegment,
		ProviderRudderstack,
		ProviderPostHog,
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		for _, provider := range allProviders {
			cfg := &Config{
				SourceConfig: SourceConfig{
					Provider:       provider,
					Segment:        &segment.Config{APIToken: t.Name()},
					Rudderstack:    &rudderstack.Config{DataPlaneURL: t.Name(), APIKey: t.Name()},
					Posthog:        &posthog.Config{APIKey: t.Name()},
					CircuitBreaker: circuitbreakingcfg.Config{},
				},
			}

			_, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
			require.NoError(t, err)
		}
	})

	T.Run("with invalid values", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		for _, provider := range allProviders {
			cfg := &Config{
				SourceConfig: SourceConfig{
					Provider:    provider,
					Segment:     &segment.Config{},
					Rudderstack: &rudderstack.Config{},
					Posthog:     &posthog.Config{},
				},
			}

			_, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
			require.Error(t, err)
		}
	})
}
