package llmcfg

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/llm/anthropic"
	"github.com/verygoodsoftwarenotvirus/platform/v5/llm/openai"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("openai provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderOpenAI,
			OpenAI: &openai.Config{
				APIKey: "test-key",
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("anthropic provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderAnthropic,
			Anthropic: &anthropic.Config{
				APIKey: "test-key",
			},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("empty provider is valid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("unknown provider is invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "nonsense",
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("openai provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderOpenAI,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("anthropic provider missing config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderAnthropic,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideLLMProvider(T *testing.T) {
	T.Parallel()

	T.Run("empty provider falls back to noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: ""}

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	T.Run("unknown provider falls back to noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: "unknown"}

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	T.Run("openai provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderOpenAI,
			OpenAI: &openai.Config{
				APIKey: "test-key",
			},
		}

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	T.Run("anthropic provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderAnthropic,
			Anthropic: &anthropic.Config{
				APIKey: "test-key",
			},
		}

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil)
		require.NoError(t, err)
		require.NotNil(t, provider)
	})

	T.Run("openai provider with metrics error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderOpenAI,
			OpenAI: &openai.Config{
				APIKey: "test-key",
			},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.AnythingOfType("string"), []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp)
		assert.Nil(t, provider)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("anthropic provider with metrics error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderAnthropic,
			Anthropic: &anthropic.Config{
				APIKey: "test-key",
			},
		}

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", mock.AnythingOfType("string"), []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		provider, err := cfg.ProvideLLMProvider(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), mp)
		assert.Nil(t, provider)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestProvideLLMProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}

		provider, err := ProvideLLMProvider(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
}
