package noop

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncNotifier_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewAsyncNotifier()
		require.NoError(t, err)
		require.NotNil(t, n)

		err = n.Publish(context.Background(), "test-channel", &async.Event{
			Type: "test",
			Data: json.RawMessage(`{"key":"value"}`),
		})
		assert.NoError(t, err)
	})
}

func TestAsyncNotifier_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewAsyncNotifier()
		require.NoError(t, err)
		require.NotNil(t, n)

		assert.NoError(t, n.Close())
	})
}
