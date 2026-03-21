package sse

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/eventstream"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpgrader(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		u := NewUpgrader()
		assert.NotNil(t, u)
	})
}

func TestUpgrader_UpgradeToEventStream(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
		assert.Equal(t, "no-cache", resp.Header.Get("Cache-Control"))
		assert.Equal(t, "keep-alive", resp.Header.Get("Connection"))
	})

	T.Run("response writer does not support flushing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		u := NewUpgrader()
		w := &nonFlushableResponseWriter{header: http.Header{}}
		r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", http.NoBody)

		stream, err := u.UpgradeToEventStream(w, r)
		assert.Nil(t, stream)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "streaming not supported")
	})
}

func TestSSEStream_Send(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		sendErr := stream.Send(t.Context(), &eventstream.Event{
			Type:    "test_event",
			Payload: json.RawMessage(`{"msg":"hello"}`),
		})
		require.NoError(t, sendErr)

		scanner := bufio.NewScanner(resp.Body)

		// Read "event: test_event"
		require.True(t, scanner.Scan())
		assert.Equal(t, "event: test_event", scanner.Text())

		// Read "data: {\"msg\":\"hello\"}"
		require.True(t, scanner.Scan())
		assert.Equal(t, `data: {"msg":"hello"}`, scanner.Text())

		// Read empty line (event separator)
		require.True(t, scanner.Scan())
		assert.Equal(t, "", scanner.Text())
	})

	T.Run("event without type", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		sendErr := stream.Send(t.Context(), &eventstream.Event{
			Payload: json.RawMessage(`{"x":1}`),
		})
		require.NoError(t, sendErr)

		scanner := bufio.NewScanner(resp.Body)

		// No "event:" line, just data
		require.True(t, scanner.Scan())
		assert.Equal(t, `data: {"x":1}`, scanner.Text())

		// Empty line (event separator)
		require.True(t, scanner.Scan())
		assert.Equal(t, "", scanner.Text())
	})

	T.Run("multiple events", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		for i, name := range []string{"first", "second", "third"} {
			sendErr := stream.Send(t.Context(), &eventstream.Event{
				Type:    "msg",
				Payload: json.RawMessage(`"` + name + `"`),
			})
			require.NoError(t, sendErr, "send %d", i)
		}

		scanner := bufio.NewScanner(resp.Body)
		for _, name := range []string{"first", "second", "third"} {
			require.True(t, scanner.Scan())
			assert.Equal(t, "event: msg", scanner.Text())

			require.True(t, scanner.Scan())
			assert.Equal(t, `data: "`+name+`"`, scanner.Text())

			require.True(t, scanner.Scan())
			assert.Equal(t, "", scanner.Text())
		}
	})

	T.Run("send after close returns error", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)

		require.NoError(t, stream.Close())

		sendErr := stream.Send(t.Context(), &eventstream.Event{
			Type:    "test",
			Payload: json.RawMessage(`{}`),
		})
		assert.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "stream closed")
	})
}

func TestSSEStream_Done(T *testing.T) {
	T.Parallel()

	T.Run("closes on Close", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)

		done := stream.Done()
		require.NoError(t, stream.Close())

		select {
		case <-done:
			// expected: channel closed
		case <-time.After(time.Second):
			require.Fail(t, "Done() channel was not closed after Close()")
		}
	})

	T.Run("closes on client disconnect", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		stream := <-streamReady
		require.NotNil(t, stream)

		// Close the client connection, which cancels the request context
		resp.Body.Close()

		// The done channel should close because the request context was cancelled
		select {
		case <-stream.Done():
			// expected
		case <-time.After(2 * time.Second):
			require.Fail(t, "Done() channel was not closed after client disconnect")
		}
	})
}

func TestSSEStream_Close(T *testing.T) {
	T.Parallel()

	T.Run("idempotent", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)

		// Close should be idempotent (context.CancelFunc is safe to call multiple times)
		assert.NoError(t, stream.Close())
		assert.NoError(t, stream.Close())
	})
}

// nonFlushableResponseWriter is a minimal ResponseWriter that does NOT implement http.Flusher.
type nonFlushableResponseWriter struct {
	header http.Header
}

func (w *nonFlushableResponseWriter) Header() http.Header         { return w.header }
func (w *nonFlushableResponseWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *nonFlushableResponseWriter) WriteHeader(int)             {}

func TestSSEStream_Send_verifies_SSE_format(T *testing.T) {
	T.Parallel()

	T.Run("output is valid SSE", func(t *testing.T) {
		t.Parallel()

		streamReady := make(chan eventstream.EventStream, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := NewUpgrader()
			stream, err := u.UpgradeToEventStream(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			streamReady <- stream
			<-stream.Done()
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		stream := <-streamReady
		require.NotNil(t, stream)
		defer stream.Close()

		sendErr := stream.Send(t.Context(), &eventstream.Event{
			Type:    "update",
			Payload: json.RawMessage(`{"id":"abc","status":"done"}`),
		})
		require.NoError(t, sendErr)

		// Read raw bytes and verify the exact SSE format
		buf := make([]byte, 4096)
		n, readErr := resp.Body.Read(buf)
		require.NoError(t, readErr)

		output := string(buf[:n])
		expected := "event: update\ndata: {\"id\":\"abc\",\"status\":\"done\"}\n\n"
		assert.Equal(t, expected, output)
	})
}
