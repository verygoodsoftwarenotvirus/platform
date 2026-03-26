package capitalismmock

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v3/capitalism"

	"github.com/stretchr/testify/mock"
)

var _ capitalism.PaymentManager = (*MockPaymentManager)(nil)

// MockPaymentManager is a mockable capitalism.PaymentManager.
type MockPaymentManager struct {
	mock.Mock
}

// NewMockPaymentManager returns a mockable capitalism.PaymentManager.
func NewMockPaymentManager() *MockPaymentManager {
	return &MockPaymentManager{}
}

// HandleEventWebhook satisfies our interface contract.
func (m *MockPaymentManager) HandleEventWebhook(req *http.Request) error {
	return m.Called(req).Error(0)
}
