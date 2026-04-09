package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/capitalism/stripe"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Enabled:  true,
			Provider: StripeProvider,
			Stripe:   &stripe.Config{APIKey: t.Name()},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("returns nil when not enabled", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Enabled: false,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Enabled:  true,
			Provider: StripeProvider,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideCapitalismImplementation(T *testing.T) {
	T.Parallel()

	T.Run("with stripe provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: StripeProvider,
			Stripe:   &stripe.Config{APIKey: t.Name()},
		}

		pm, err := ProvideCapitalismImplementation(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cfg)
		require.NoError(t, err)
		assert.NotNil(t, pm)
	})

	T.Run("with unknown provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: "unknown",
		}

		pm, err := ProvideCapitalismImplementation(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), cfg)
		assert.Nil(t, pm)
		assert.Error(t, err)
	})
}
