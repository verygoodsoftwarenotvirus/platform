package emailcfg

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email/mailgun"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email/mailjet"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email/postmark"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email/resend"
	"github.com/verygoodsoftwarenotvirus/platform/v2/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/matcornic/hermes/v2"
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
		Sendgrid                            *sendgrid.Config       `env:"init"                                    envPrefix:"SENDGRID_"                      json:"sendgrid"`
		Mailgun                             *mailgun.Config        `env:"init"                                    envPrefix:"MAILGUN_"                       json:"mailgun"`
		Mailjet                             *mailjet.Config        `env:"init"                                    envPrefix:"MAILJET_"                       json:"mailjet"`
		Resend                              *resend.Config         `env:"init"                                    envPrefix:"RESEND_"                        json:"resend"`
		Postmark                            *postmark.Config       `env:"init"                                    envPrefix:"POSTMARK_"                      json:"postmark"`
		Provider                            string                 `env:"PROVIDER"                                json:"provider"`
		BaseURL                             template.URL           `env:"BASE_URL"                                json:"baseURL"`
		OutboundInvitesEmailAddress         string                 `env:"OUTBOUND_INVITES_EMAIL_ADDRESS"          json:"outboundInvitesEmailAddress"`
		PasswordResetCreationEmailAddress   string                 `env:"PASSWORD_RESET_CREATION_EMAIL_ADDRESS"   json:"passwordResetCreationEmailAddress"`
		PasswordResetRedemptionEmailAddress string                 `env:"PASSWORD_RESET_REDEMPTION_EMAIL_ADDRESS" json:"passwordResetRedemptionEmailAddress"`
		CircuitBreaker                      circuitbreaking.Config `env:"init"                                    envPrefix:"CIRCUIT_BREAKING_"              json:"circuitBreakerConfig"`
	}
)

// BuildHermes builds a Hermes instance for rendering email templates.
func (cfg *Config) BuildHermes(branding *email.EmailBranding) *hermes.Hermes {
	var name, logo, copyright string
	if branding != nil {
		name = branding.CompanyName
		logo = branding.LogoURL
		copyright = fmt.Sprintf("Copyright © %d %s. All rights reserved.", time.Now().Year(), branding.CompanyName)
	}
	return &hermes.Hermes{
		Product: hermes.Product{
			Name:      name,
			Link:      string(cfg.BaseURL),
			Logo:      logo,
			Copyright: copyright,
		},
	}
}

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
