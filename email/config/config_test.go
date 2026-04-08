package emailcfg

import (
	"fmt"
	"net/http"
	"testing"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/mailgun"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/mailjet"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/postmark"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/resend"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email/ses"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

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

	T.Run("mailgun provider requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderMailgun}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("mailjet provider requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderMailjet}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("resend provider requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderResend}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("postmark provider requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderPostmark}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("ses provider requires config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderSES}
		require.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestConfig_BuildHermes(T *testing.T) {
	T.Parallel()

	T.Run("with branding", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{BaseURL: "https://example.com"}
		h := cfg.BuildHermes(&email.EmailBranding{
			CompanyName: "Acme",
			LogoURL:     "https://example.com/logo.png",
		})
		require.NotNil(t, h)
		assert.Equal(t, "Acme", h.Product.Name)
		assert.Equal(t, "https://example.com/logo.png", h.Product.Logo)
		assert.Equal(t, "https://example.com", h.Product.Link)
		assert.Contains(t, h.Product.Copyright, "Acme")
	})

	T.Run("without branding", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{BaseURL: "https://example.com"}
		h := cfg.BuildHermes(nil)
		require.NotNil(t, h)
		assert.Empty(t, h.Product.Name)
		assert.Empty(t, h.Product.Logo)
		assert.Empty(t, h.Product.Copyright)
	})
}

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()
		assert.NotEmpty(t, cfg.CircuitBreaker.Name)
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

			actual, err := cfg.ProvideEmailer(t.Context(), logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
			assert.NotNil(t, actual)
			assert.NoError(t, err)
		})
	}

	T.Run("with ses provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: ProviderSES,
			SES:      &ses.Config{Region: "us-east-1"},
		}

		actual, err := cfg.ProvideEmailer(t.Context(), logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: "",
		}

		actual, err := cfg.ProvideEmailer(t.Context(), logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}

func TestProvideEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard falls back to noop", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.CircuitBreaker.Name = t.Name()

		emailer, err := ProvideEmailer(
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			&http.Client{},
		)
		require.NoError(t, err)
		assert.NotNil(t, emailer)
	})

	T.Run("with sendgrid provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderSendgrid,
			Sendgrid: &sendgrid.Config{APIToken: t.Name()},
		}
		cfg.CircuitBreaker.Name = t.Name()

		emailer, err := ProvideEmailer(
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			metrics.NewNoopMetricsProvider(),
			&http.Client{},
		)
		require.NoError(t, err)
		assert.NotNil(t, emailer)
	})

	T.Run("circuit breaker init failure", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.CircuitBreaker.Name = "email-breaker"
		cfg.CircuitBreaker.ErrorRate = 50
		cfg.CircuitBreaker.MinimumSampleThreshold = 10

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", "email-breaker_circuit_breaker_tripped", []metric.Int64CounterOption(nil)).
			Return(&mockmetrics.Int64Counter{}, fmt.Errorf("counter init failure"))

		emailer, err := ProvideEmailer(
			t.Context(),
			cfg,
			logging.NewNoopLogger(),
			tracing.NewNoopTracerProvider(),
			mp,
			&http.Client{},
		)
		require.Error(t, err)
		assert.Nil(t, emailer)
		mp.AssertExpectations(t)
	})
}
