package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		emailer, err := NewNoopEmailer()
		require.NoError(t, err)
		assert.NotNil(t, emailer)
	})
}

func TestNoopEmailer_SendEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		emailer, err := NewNoopEmailer()
		require.NoError(t, err)

		err = emailer.SendEmail(context.Background(), &OutboundEmailMessage{
			ToAddress:   "test@example.com",
			Subject:     "Test",
			HTMLContent: "<p>hello</p>",
		})
		assert.NoError(t, err)
	})

	T.Run("with nil message", func(t *testing.T) {
		t.Parallel()

		emailer, err := NewNoopEmailer()
		require.NoError(t, err)

		err = emailer.SendEmail(context.Background(), nil)
		assert.NoError(t, err)
	})
}
