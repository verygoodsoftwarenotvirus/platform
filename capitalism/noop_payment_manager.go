package capitalism

import (
	"net/http"
)

var _ PaymentManager = (*NoopPaymentManager)(nil)

// NoopPaymentManager is a no-op payment manager.
type NoopPaymentManager struct{}

// HandleEventWebhook satisfies our interface.
func (n *NoopPaymentManager) HandleEventWebhook(_ *http.Request) error {
	return nil
}
