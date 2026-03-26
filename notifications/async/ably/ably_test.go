package ably

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChannelPublisher struct {
	publishFn func(ctx context.Context, channel, name string, data any) error
}

func (m *mockChannelPublisher) Publish(ctx context.Context, channel, name string, data any) error {
	return m.publishFn(ctx, channel, name, data)
}

func TestNewNotifier(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{
			APIKey: "appid.keyid:keysecret",
		}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		require.NoError(t, err)
		require.NotNil(t, n)
	})

	T.Run("nil config", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		assert.Error(t, err)
		assert.Nil(t, n)
	})
}

func TestNotifier_Publish(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var capturedChannel, capturedName string
		n := &Notifier{
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest("test"),
			publisher: &mockChannelPublisher{
				publishFn: func(_ context.Context, channel, name string, _ any) error {
					capturedChannel = channel
					capturedName = name
					return nil
				},
			},
		}

		err := n.Publish(context.Background(), "my-channel", &async.Event{
			Type: "greeting",
			Data: json.RawMessage(`{"hello":"world"}`),
		})
		assert.NoError(t, err)
		assert.Equal(t, "my-channel", capturedChannel)
		assert.Equal(t, "greeting", capturedName)
	})

	T.Run("publish error", func(t *testing.T) {
		t.Parallel()

		n := &Notifier{
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest("test"),
			publisher: &mockChannelPublisher{
				publishFn: func(context.Context, string, string, any) error {
					return errors.New("ably API error")
				},
			},
		}

		err := n.Publish(context.Background(), "my-channel", &async.Event{
			Type: "test",
		})
		assert.Error(t, err)
	})
}

func TestNotifier_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n := &Notifier{
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracerForTest("test"),
		}

		assert.NoError(t, n.Close())
	})
}
