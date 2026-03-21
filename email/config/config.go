package emailcfg

import (
	"context"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/email"
	"github.com/verygoodsoftwarenotvirus/platform/email/mailgun"
	"github.com/verygoodsoftwarenotvirus/platform/email/mailjet"
	"github.com/verygoodsoftwarenotvirus/platform/email/postmark"
	"github.com/verygoodsoftwarenotvirus/platform/email/resend"
	"github.com/verygoodsoftwarenotvirus/platform/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderSendgrid represents SendGrid.
	ProviderSendgrid = "sendgrid"
	// ProviderMailgun represents Mailgun.
	ProviderMailgun = "mailgun"
	// ProviderMailjet represents Mailjet.
	ProviderMailjet = "mailjet"
	// ProviderResend represents Resend.
	ProviderResend = "resend"
	// ProviderPostmark represents Postmark.
	ProviderPostmark = "postmark"
)

type (
	// Config is the configuration structure.
	Config struct {
		Sendgrid       *sendgrid.Config       `env:"init"     envPrefix:"SENDGRID_"         json:"sendgrid"`
		Mailgun        *mailgun.Config        `env:"init"     envPrefix:"MAILGUN_"          json:"mailgun"`
		Mailjet        *mailjet.Config        `env:"init"     envPrefix:"MAILJET_"          json:"mailjet"`
		Resend         *resend.Config         `env:"init"     envPrefix:"RESEND_"           json:"resend"`
		Postmark       *postmark.Config       `env:"init"     envPrefix:"POSTMARK_"         json:"postmark"`
		Provider       string                 `env:"PROVIDER" json:"provider"`
		CircuitBreaker circuitbreaking.Config `env:"init"     envPrefix:"CIRCUIT_BREAKING_" json:"circuitBreakerConfig"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Sendgrid, validation.When(cfg.Provider == ProviderSendgrid, validation.Required)),
		validation.Field(&cfg.Mailgun, validation.When(cfg.Provider == ProviderMailgun, validation.Required)),
		validation.Field(&cfg.Mailjet, validation.When(cfg.Provider == ProviderMailjet, validation.Required)),
		validation.Field(&cfg.Resend, validation.When(cfg.Provider == ProviderResend, validation.Required)),
		validation.Field(&cfg.Postmark, validation.When(cfg.Provider == ProviderPostmark, validation.Required)),
	)
}

// ProvideEmailer provides an outbound_emailer.
func (cfg *Config) ProvideEmailer(logger logging.Logger, tracerProvider tracing.TracerProvider, client *http.Client, circuitBreaker circuitbreaking.CircuitBreaker) (email.Emailer, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case ProviderSendgrid:
		return sendgrid.NewSendGridEmailer(cfg.Sendgrid, logger, tracerProvider, client, circuitBreaker)
	case ProviderMailgun:
		return mailgun.NewMailgunEmailer(cfg.Mailgun, logger, tracerProvider, client, circuitBreaker)
	case ProviderMailjet:
		return mailjet.NewMailjetEmailer(cfg.Mailjet, logger, tracerProvider, client, circuitBreaker)
	case ProviderResend:
		return resend.NewResendEmailer(cfg.Resend, logger, tracerProvider, client, circuitBreaker)
	case ProviderPostmark:
		return postmark.NewPostmarkEmailer(cfg.Postmark, logger, tracerProvider, client, circuitBreaker)
	default:
		logger.Debug("providing noop outbound_emailer")
		return email.NewNoopEmailer()
	}
}
