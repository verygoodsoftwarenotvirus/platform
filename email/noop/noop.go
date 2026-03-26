package noop

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/email"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
)

var _ email.Emailer = (*emailer)(nil)

// emailer doesn't send emails.
type emailer struct {
	logger logging.Logger
}

// NewEmailer returns a new no-op Emailer.
func NewEmailer() (email.Emailer, error) {
	return &emailer{logger: logging.NewNoopLogger()}, nil
}

// SendEmail sends an email.
func (e *emailer) SendEmail(context.Context, *email.OutboundEmailMessage) error {
	e.logger.Info("NoopEmailer.SendEmail: no-op")
	return nil
}
