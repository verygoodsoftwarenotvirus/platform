package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotifier(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
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

	T.Run("no connected clients", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		require.NoError(t, err)

		err = n.Publish(context.Background(), "test-channel", &async.Event{
			Type: "test",
			Data: json.RawMessage(`{"key":"value"}`),
		})
		assert.NoError(t, err)
	})

	T.Run("with connected client", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptErr := n.AcceptConnection(w, r, "test-channel", "member-1")
			assert.NoError(t, acceptErr)
			// keep the handler alive so the websocket stays open
			<-r.Context().Done()
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// give the connection time to register
		time.Sleep(50 * time.Millisecond)

		err = n.Publish(context.Background(), "test-channel", &async.Event{
			Type: "greeting",
			Data: json.RawMessage(`{"hello":"world"}`),
		})
		assert.NoError(t, err)

		var received map[string]json.RawMessage
		err = conn.ReadJSON(&received)
		require.NoError(t, err)
		assert.Equal(t, json.RawMessage(`"greeting"`), received["type"])
	})
}

func TestNotifier_AcceptConnection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		require.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptErr := n.AcceptConnection(w, r, "channel", "member")
			assert.NoError(t, acceptErr)
			<-r.Context().Done()
		}))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()
	})
}

func TestNotifier_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		n, err := NewNotifier(&Config{}, logging.NewNoopLogger(), tracing.NewNoopTracerProvider())
		require.NoError(t, err)

		assert.NoError(t, n.Close())
	})
}
