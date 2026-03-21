package emailcfg

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/email/mailgun"
	"github.com/verygoodsoftwarenotvirus/platform/email/mailjet"
	"github.com/verygoodsoftwarenotvirus/platform/email/postmark"
	"github.com/verygoodsoftwarenotvirus/platform/email/resend"
	"github.com/verygoodsoftwarenotvirus/platform/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Sendgrid: &sendgrid.Config{APIToken: t.Name()},
		}

		require.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "sendgrid",
		}

		require.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideEmailer(T *testing.T) {
	T.Parallel()

	providers := []string{
		ProviderSendgrid,
		ProviderMailgun,
		ProviderMailjet,
		ProviderResend,
		ProviderPostmark,
	}

	for _, provider := range providers {
		T.Run(fmt.Sprintf("with %s", provider), func(t *testing.T) {
			t.Parallel()

			logger := logging.NewNoopLogger()
			cfg := &Config{
				Provider: provider,
				Sendgrid: &sendgrid.Config{APIToken: t.Name()},
				Mailgun:  &mailgun.Config{PrivateAPIKey: t.Name(), Domain: t.Name()},
				Mailjet:  &mailjet.Config{APIKey: t.Name(), SecretKey: t.Name()},
				Resend:   &resend.Config{APIToken: t.Name()},
				Postmark: &postmark.Config{ServerToken: t.Name()},
			}

			actual, err := cfg.ProvideEmailer(logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
			assert.NotNil(t, actual)
			assert.NoError(t, err)
		})
	}

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: "",
		}

		actual, err := cfg.ProvideEmailer(logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
