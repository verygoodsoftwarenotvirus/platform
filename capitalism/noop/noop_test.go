package noop

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentManager_HandleEventWebhook(T *testing.T) {
	T.Parallel()

	T.Run("returns nil", func(t *testing.T) {
		t.Parallel()
		mgr := NewPaymentManager()
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "https://example.com/webhook", http.NoBody)
		require.NoError(t, err)

		assert.NoError(t, mgr.HandleEventWebhook(req))
	})
}

func TestPaymentManager_ImplementsInterface(T *testing.T) {
	T.Parallel()

	T.Run("satisfies PaymentManager", func(t *testing.T) {
		t.Parallel()
		_ = NewPaymentManager()
	})
}
