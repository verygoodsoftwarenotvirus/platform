package email

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
)

var _ Emailer = (*NoopEmailer)(nil)

type (
	// NoopEmailer doesn't send emails.
	NoopEmailer struct {
		logger logging.Logger
	}
)

// NewNoopEmailer returns a new no-op NoopEmailer.
func NewNoopEmailer() (*NoopEmailer, error) {
	return &NoopEmailer{logger: logging.NewNoopLogger()}, nil
}

// SendEmail sends an email.
func (e *NoopEmailer) SendEmail(context.Context, *OutboundEmailMessage) error {
	e.logger.Info("NoopEmailer.SendEmail: no-op")
	return nil
}
