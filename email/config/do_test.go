package emailcfg

import (
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/email"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderSendgrid,
			Sendgrid: &sendgrid.Config{APIToken: t.Name()},
		}
		cfg.CircuitBreaker.Name = t.Name()

		i := do.New()
		do.ProvideValue(i, t.Context())
		do.ProvideValue(i, logging.NewNoopLogger())
		do.ProvideValue(i, tracing.NewNoopTracerProvider())
		do.ProvideValue[metrics.Provider](i, metrics.NewNoopMetricsProvider())
		do.ProvideValue(i, &http.Client{})
		do.ProvideValue(i, cfg)

		RegisterEmailer(i)

		emailer, err := do.Invoke[email.Emailer](i)
		require.NoError(t, err)
		assert.NotNil(t, emailer)
	})
}
