package pusher

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

type mockPusherClient struct {
	triggerFn func(channel, eventName string, data any) error
}

func (m *mockPusherClient) Trigger(channel, eventName string, data any) error {
	return m.triggerFn(channel, eventName, data)
}

func TestNewNotifier(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{
			AppID:   "123",
			Key:     "key",
			Secret:  "secret",
			Cluster: "us2",
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

		var capturedChannel, capturedEvent string
		n := &Notifier{
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test")),
			client: &mockPusherClient{
				triggerFn: func(channel, eventName string, data any) error {
					capturedChannel = channel
					capturedEvent = eventName
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
		assert.Equal(t, "greeting", capturedEvent)
	})

	T.Run("trigger error", func(t *testing.T) {
		t.Parallel()

		n := &Notifier{
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test")),
			client: &mockPusherClient{
				triggerFn: func(string, string, any) error {
					return errors.New("pusher API error")
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
			tracer: tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test")),
		}

		assert.NoError(t, n.Close())
	})
}
