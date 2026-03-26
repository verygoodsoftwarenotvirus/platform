package postmark

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v3/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v3/email"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v3/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/keighl/postmark"
)

const (
	name = "postmark_emailer"
)

var (
	_ email.Emailer = (*Emailer)(nil)

	// ErrNilConfig indicates a nil config was provided.
	ErrNilConfig = platformerrors.New("postmark config is nil")
	// ErrEmptyServerToken indicates an empty server token was provided.
	ErrEmptyServerToken = platformerrors.New("empty Postmark server token")
	// ErrNilHTTPClient indicates a nil HTTP client was provided.
	ErrNilHTTPClient = platformerrors.New("nil HTTP client")
)

type (
	// Emailer uses Postmark to send email.
	Emailer struct {
		logger         logging.Logger
		tracer         tracing.Tracer
		client         *postmark.Client
		circuitBreaker circuitbreaking.CircuitBreaker
	}
)

// NewPostmarkEmailer returns a new Postmark-backed Emailer.
func NewPostmarkEmailer(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider, client *http.Client, circuitBreaker circuitbreaking.CircuitBreaker) (*Emailer, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	if strings.TrimSpace(cfg.ServerToken) == "" {
		return nil, ErrEmptyServerToken
	}

	if client == nil {
		return nil, ErrNilHTTPClient
	}

	pm := postmark.NewClient(cfg.ServerToken, "")
	pm.HTTPClient = client
	if cfg.BaseURL != "" {
		pm.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	}

	e := &Emailer{
		logger:         logging.EnsureLogger(logger).WithName(name),
		tracer:         tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(name)),
		client:         pm,
		circuitBreaker: circuitBreaker,
	}

	return e, nil
}

func formatAddress(name, address string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", name, address)
}

// SendEmail sends an email.
func (e *Emailer) SendEmail(ctx context.Context, details *email.OutboundEmailMessage) error {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	logger := e.logger.WithValue("email.subject", details.Subject).WithValue("email.to_address", details.ToAddress)
	tracing.AttachToSpan(span, "to_email", details.ToAddress)

	if e.circuitBreaker.CannotProceed() {
		return circuitbreaking.ErrCircuitBroken
	}

	pmEmail := postmark.Email{
		From:     formatAddress(details.FromName, details.FromAddress),
		To:       formatAddress(details.ToName, details.ToAddress),
		Subject:  details.Subject,
		HtmlBody: details.HTMLContent,
	}

	if _, err := e.client.SendEmail(pmEmail); err != nil {
		e.circuitBreaker.Failed()
		return observability.PrepareAndLogError(err, logger, span, "sending email")
	}

	e.circuitBreaker.Succeeded()
	return nil
}
