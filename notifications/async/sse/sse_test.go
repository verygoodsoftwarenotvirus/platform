package sse

import (
	"bufio"
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
		require.NoError(t, err)
		require.NotNil(t, n)
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

		ready := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(ready)
			acceptErr := n.AcceptConnection(w, r, "test-channel", "member-1")
			assert.NoError(t, acceptErr)
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		<-ready
		// give the connection time to register
		time.Sleep(50 * time.Millisecond)

		err = n.Publish(context.Background(), "test-channel", &async.Event{
			Type: "greeting",
			Data: json.RawMessage(`{"hello":"world"}`),
		})
		assert.NoError(t, err)

		scanner := bufio.NewScanner(resp.Body)
		var eventLine, dataLine string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				eventLine = line
			}
			if strings.HasPrefix(line, "data:") {
				dataLine = line
				break
			}
		}

		assert.Contains(t, eventLine, "greeting")
		assert.Contains(t, dataLine, `{"hello":"world"}`)
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
