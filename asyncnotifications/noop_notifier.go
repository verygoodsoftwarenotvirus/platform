package asyncnotifications

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
)

var _ AsyncNotifier = (*NoopAsyncNotifier)(nil)

// NoopAsyncNotifier is a no-op implementation of AsyncNotifier.
type NoopAsyncNotifier struct {
	logger logging.Logger
}

// NewNoopAsyncNotifier returns a new no-op NoopAsyncNotifier.
func NewNoopAsyncNotifier() (*NoopAsyncNotifier, error) {
	return &NoopAsyncNotifier{logger: logging.NewNoopLogger()}, nil
}

// Publish is a no-op.
func (n *NoopAsyncNotifier) Publish(context.Context, string, *Event) error {
	n.logger.Info("NoopAsyncNotifier.Publish: no-op")
	return nil
}

// Close is a no-op.
func (n *NoopAsyncNotifier) Close() error {
	return nil
}
