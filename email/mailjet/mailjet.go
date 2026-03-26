package mailjet

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v3/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v3/email"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/mailjet/mailjet-apiv3-go/v4"
)

const (
	name = "mailjet_emailer"
)

var (
	_ email.Emailer = (*Emailer)(nil)

	// ErrNilConfig indicates a nil config was provided.
	ErrNilConfig = platformerrors.New("mailjet config is nil")
	// ErrEmptySecretKey indicates an empty domain was provided.
	ErrEmptySecretKey = platformerrors.New("empty domain")
	// ErrEmptyPrivateAPIKey indicates an empty API token was provided.
	ErrEmptyPrivateAPIKey = platformerrors.New("empty Mailjet API token")
	// ErrNilHTTPClient indicates a nil HTTP client was provided.
	ErrNilHTTPClient = platformerrors.New("nil HTTP client")
)

type (
	mailjetClient interface {
		SendMailV31(data *mailjet.MessagesV31, options ...mailjet.RequestOptions) (*mailjet.ResultsV31, error)
	}

	// Emailer uses Mailjet to send email.
	Emailer struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		client         mailjetClient
		circuitBreaker circuitbreaking.CircuitBreaker
	}
)

// NewMailjetEmailer returns a new Mailjet-backed Emailer.
func NewMailjetEmailer(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, client *http.Client, circuitBreaker circuitbreaking.CircuitBreaker) (*Emailer, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	if cfg.SecretKey == "" {
		return nil, ErrEmptySecretKey
	}

	if cfg.APIKey == "" {
		return nil, ErrEmptyPrivateAPIKey
	}

	if client == nil {
		return nil, ErrNilHTTPClient
	}

	mj := mailjet.NewMailjetClient(cfg.APIKey, cfg.SecretKey)
	mj.SetClient(client)

	e := &Emailer{
		logger:         logging.EnsureLogger(logger).WithName(name),
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(name)),
		client:         mj,
		circuitBreaker: circuitBreaker,
	}

	return e, nil
}

// SendEmail sends an email.
func (e *Emailer) SendEmail(ctx context.Context, details *email.OutboundEmailMessage) error {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	if e.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	logger := e.logger.WithValue("email.subject", details.Subject).WithValue("email.to_address", details.ToAddress)
	tracing.AttachToSpan(span, "to_email", details.ToAddress)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: details.FromAddress,
				Name:  details.FromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: details.ToAddress,
					Name:  details.ToName,
				},
			},
			Subject:  details.Subject,
			HTMLPart: details.HTMLContent,
		},
	}

	if _, err := e.client.SendMailV31(&mailjet.MessagesV31{Info: messagesInfo}); err != nil {
		e.circuitBreaker.Failed()
		return observability.PrepareAndLogError(err, logger, span, "sending email")
	}

	e.circuitBreaker.Succeeded()
	return nil
}
