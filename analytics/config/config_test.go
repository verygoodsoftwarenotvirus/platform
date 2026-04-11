package analyticscfg

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/posthog"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/rudderstack"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/segment"
	circuitbreakingcfg "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/shoenig/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
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

	T.Run("with segment provider but nil segment config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderSegment,
			},
		}

		reporter, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		assert.Nil(t, reporter)
		assert.Error(t, err)
	})

	T.Run("with rudderstack provider but nil rudderstack config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderRudderstack,
			},
		}

		reporter, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		assert.Nil(t, reporter)
		assert.Error(t, err)
	})

	T.Run("with posthog provider but nil posthog config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderPostHog,
			},
		}

		reporter, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		assert.Nil(t, reporter)
		assert.Error(t, err)
	})

	T.Run("with unrecognized provider returns noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: "bogus",
			},
		}

		reporter, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, reporter)
		assert.NoError(t, err)
	})

	T.Run("with circuit breaker error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			SourceConfig: SourceConfig{
				Provider: ProviderSegment,
				Segment:  &segment.Config{APIToken: t.Name()},
				CircuitBreaker: circuitbreakingcfg.Config{
					Name:                   t.Name(),
					ErrorRate:              99,
					MinimumSampleThreshold: 1,
				},
			},
		}

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(_ string, options ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				test.SliceEmpty(t, options)
				return nil, errors.New("arbitrary")
			},
		}

		reporter, err := cfg.ProvideCollector(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp)
		assert.Nil(t, reporter)
		assert.Error(t, err)

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})
}

func TestSourceConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &SourceConfig{}
		cfg.EnsureDefaults()

		assert.NotEmpty(t, cfg.CircuitBreaker.Name)
		assert.NotZero(t, cfg.CircuitBreaker.ErrorRate)
		assert.NotZero(t, cfg.CircuitBreaker.MinimumSampleThreshold)
	})
}

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("with nil proxy sources", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()

		assert.NotEmpty(t, cfg.CircuitBreaker.Name)
	})

	T.Run("with both proxy sources set", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			ProxySources: ProxySourcesConfig{
				IOS: &SourceConfig{},
				Web: &SourceConfig{},
			},
		}
		cfg.EnsureDefaults()

		assert.NotEmpty(t, cfg.CircuitBreaker.Name)
		assert.NotEmpty(t, cfg.ProxySources.IOS.CircuitBreaker.Name)
		assert.NotEmpty(t, cfg.ProxySources.Web.CircuitBreaker.Name)
	})
}

func TestProxySourcesConfig_ToMap(T *testing.T) {
	T.Parallel()

	T.Run("with nil sources", func(t *testing.T) {
		t.Parallel()

		p := ProxySourcesConfig{}
		assert.Empty(t, p.ToMap())
	})

	T.Run("with only ios set", func(t *testing.T) {
		t.Parallel()

		ios := &SourceConfig{Provider: ProviderSegment}
		p := ProxySourcesConfig{IOS: ios}
		m := p.ToMap()

		assert.Len(t, m, 1)
		assert.Same(t, ios, m["ios"])
	})

	T.Run("with only web set", func(t *testing.T) {
		t.Parallel()

		web := &SourceConfig{Provider: ProviderPostHog}
		p := ProxySourcesConfig{Web: web}
		m := p.ToMap()

		assert.Len(t, m, 1)
		assert.Same(t, web, m["web"])
	})

	T.Run("with both sources set", func(t *testing.T) {
		t.Parallel()

		ios := &SourceConfig{Provider: ProviderSegment}
		web := &SourceConfig{Provider: ProviderPostHog}
		p := ProxySourcesConfig{IOS: ios, Web: web}
		m := p.ToMap()

		assert.Len(t, m, 2)
		assert.Same(t, ios, m["ios"])
		assert.Same(t, web, m["web"])
	})
}
