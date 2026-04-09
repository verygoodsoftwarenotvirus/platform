package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/capitalism"
	"github.com/verygoodsoftwarenotvirus/platform/v5/capitalism/stripe"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterPaymentManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue(i, &Config{
			Provider: StripeProvider,
			Stripe:   &stripe.Config{APIKey: t.Name()},
		})

		RegisterPaymentManager(i)

		pm, err := do.Invoke[capitalism.PaymentManager](i)
		require.NoError(t, err)
		assert.NotNil(t, pm)
	})
}
