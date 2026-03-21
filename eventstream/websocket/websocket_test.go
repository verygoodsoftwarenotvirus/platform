package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/eventstream"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("nil config uses defaults", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(nil)
		require.NotNil(t, u)
		assert.Equal(t, defaultHeartbeatInterval, u.heartbeatInterval)
		assert.Equal(t, defaultBufferSize, u.wsUpgrader.ReadBufferSize)
		assert.Equal(t, defaultBufferSize, u.wsUpgrader.WriteBufferSize)
	})

	T.Run("custom config", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{
			HeartbeatInterval: 10 * time.Second,
			ReadBufferSize:    2048,
			WriteBufferSize:   4096,
		})
		require.NotNil(t, u)
		assert.Equal(t, 10*time.Second, u.heartbeatInterval)
		assert.Equal(t, 2048, u.wsUpgrader.ReadBufferSize)
		assert.Equal(t, 4096, u.wsUpgrader.WriteBufferSize)
	})
}

func TestUpgrader_UpgradeToEventStream(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()
	})
}

func TestUpgrader_UpgradeToBidirectionalStream(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.BidirectionalEventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToBidirectionalStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		assert.NotNil(t, stream.Receive())
	})
}

func TestWSStream_Send(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		received := make(chan *eventstream.Event, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				return
			}
			defer stream.Close()

			_ = stream.Send(r.Context(), &eventstream.Event{
				Type:    "test",
				Payload: json.RawMessage(`{"msg":"hello"}`),
			})
			// keep alive briefly so client can read
			time.Sleep(100 * time.Millisecond)
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		go func() {
			var event eventstream.Event
			if readErr := conn.ReadJSON(&event); readErr == nil {
				received <- &event
			}
		}()

		select {
		case event := <-received:
			assert.Equal(t, "test", event.Type)
			assert.JSONEq(t, `{"msg":"hello"}`, string(event.Payload))
		case <-time.After(2 * time.Second):
			require.Fail(t, "did not receive event")
		}
	})

	T.Run("send after close returns error", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		require.NoError(t, stream.Close())

		sendErr := stream.Send(t.Context(), &eventstream.Event{Type: "x"})
		assert.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "stream closed")
	})
}

func TestWSStream_Done(T *testing.T) {
	T.Parallel()

	T.Run("closes on Close", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		done := stream.Done()
		require.NoError(t, stream.Close())

		select {
		case <-done:
			// expected
		case <-time.After(time.Second):
			require.Fail(t, "Done() channel was not closed after Close()")
		}
	})
}

func TestWSStream_Close(T *testing.T) {
	T.Parallel()

	T.Run("idempotent", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		assert.NoError(t, stream.Close())
		assert.NoError(t, stream.Close())
	})
}

func TestBidirectionalWSStream_Receive(T *testing.T) {
	T.Parallel()

	T.Run("receives client messages", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.BidirectionalEventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToBidirectionalStream(w, r)
			if err != nil {
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		defer stream.Close()

		// Client sends an event
		outgoing := &eventstream.Event{
			Type:    "ping",
			Payload: json.RawMessage(`{"seq":1}`),
		}
		require.NoError(t, conn.WriteJSON(outgoing))

		select {
		case event := <-stream.Receive():
			require.NotNil(t, event)
			assert.Equal(t, "ping", event.Type)
			assert.JSONEq(t, `{"seq":1}`, string(event.Payload))
		case <-time.After(2 * time.Second):
			require.Fail(t, "did not receive event from client")
		}
	})

	T.Run("channel closes when stream is closed", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader(&Config{HeartbeatInterval: time.Hour})
		streamReady := make(chan eventstream.BidirectionalEventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stream, err := u.UpgradeToBidirectionalStream(w, r)
			if err != nil {
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+server.URL[4:], http.Header{"Origin": {server.URL}})
		require.NoError(t, err)
		defer conn.Close()

		stream := <-streamReady
		incoming := stream.Receive()

		require.NoError(t, stream.Close())

		select {
		case _, open := <-incoming:
			assert.False(t, open, "Receive channel should be closed")
		case <-time.After(2 * time.Second):
			require.Fail(t, "Receive channel was not closed after stream.Close()")
		}
	})
}
