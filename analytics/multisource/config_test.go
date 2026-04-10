package multisource

import (
	"testing"

	analyticscfg "github.com/verygoodsoftwarenotvirus/platform/v5/analytics/config"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/posthog"
	"github.com/verygoodsoftwarenotvirus/platform/v5/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideMultiSourceEventReporter(T *testing.T) {
	T.Parallel()

	T.Run("with no proxy sources", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		reporter, err := ProvideMultiSourceEventReporter(ctx, nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Empty(t, reporter.reporters)
	})

	T.Run("with valid segment source", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sources := map[string]*analyticscfg.SourceConfig{
			"ios": {
				Provider: analyticscfg.ProviderSegment,
				Segment:  &segment.Config{APIToken: t.Name()},
			},
		}

		reporter, err := ProvideMultiSourceEventReporter(ctx, sources, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Len(t, reporter.reporters, 1)
	})

	T.Run("with invalid source falls back to noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sources := map[string]*analyticscfg.SourceConfig{
			"ios": {
				Provider: analyticscfg.ProviderSegment,
				Segment:  &segment.Config{},
			},
		}

		reporter, err := ProvideMultiSourceEventReporter(ctx, sources, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Len(t, reporter.reporters, 1)
	})

	T.Run("with unrecognized provider uses noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sources := map[string]*analyticscfg.SourceConfig{
			"web": {
				Provider: "bogus",
			},
		}

		reporter, err := ProvideMultiSourceEventReporter(ctx, sources, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Len(t, reporter.reporters, 1)
	})

	T.Run("with multiple posthog sources reuses shared reporter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sources := map[string]*analyticscfg.SourceConfig{
			"ios": {
				Provider: analyticscfg.ProviderPostHog,
				Posthog:  &posthog.Config{APIKey: t.Name()},
			},
			"web": {
				Provider: analyticscfg.ProviderPostHog,
				Posthog:  &posthog.Config{APIKey: t.Name()},
			},
		}

		reporter, err := ProvideMultiSourceEventReporter(ctx, sources, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Len(t, reporter.reporters, 2)
	})

	T.Run("with empty proxy sources map", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sources := map[string]*analyticscfg.SourceConfig{}

		reporter, err := ProvideMultiSourceEventReporter(ctx, sources, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider())
		require.NoError(t, err)
		require.NotNil(t, reporter)
		assert.Empty(t, reporter.reporters)
	})
}
