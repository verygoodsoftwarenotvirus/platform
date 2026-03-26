package noop

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		e, err := NewEmailer()
		require.NoError(t, err)
		assert.NotNil(t, e)
	})
}

func TestEmailer_SendEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		e, err := NewEmailer()
		require.NoError(t, err)

		err = e.SendEmail(context.Background(), &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			Subject:     "Test",
			HTMLContent: "<p>hello</p>",
		})
		assert.NoError(t, err)
	})

	T.Run("with nil message", func(t *testing.T) {
		t.Parallel()

		e, err := NewEmailer()
		require.NoError(t, err)

		err = e.SendEmail(context.Background(), nil)
		assert.NoError(t, err)
	})
}
