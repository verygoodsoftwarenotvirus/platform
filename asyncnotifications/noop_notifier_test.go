package asyncnotifications

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoopAsyncNotifier_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNoopAsyncNotifier()
		require.NoError(t, err)
		require.NotNil(t, n)

		err = n.Publish(context.Background(), "test-channel", &Event{
			Type: "test",
			Data: json.RawMessage(`{"key":"value"}`),
		})
		assert.NoError(t, err)
	})
}

func TestNoopAsyncNotifier_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNoopAsyncNotifier()
		require.NoError(t, err)
		require.NotNil(t, n)

		assert.NoError(t, n.Close())
	})
}
