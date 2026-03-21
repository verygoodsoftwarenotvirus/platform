package sse

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/verygoodsoftwarenotvirus/platform/eventstream"
)

var (
	_ eventstream.EventStreamUpgrader = (*Upgrader)(nil)
	_ eventstream.EventStream         = (*sseStream)(nil)
)

// Upgrader upgrades HTTP connections to SSE event streams.
type Upgrader struct{}

// NewUpgrader creates a new SSE Upgrader.
func NewUpgrader() *Upgrader {
	return &Upgrader{}
}

// UpgradeToEventStream upgrades an HTTP connection to a unidirectional SSE event stream.
func (u *Upgrader) UpgradeToEventStream(w http.ResponseWriter, r *http.Request) (eventstream.EventStream, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported by response writer")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	ctx, cancel := context.WithCancel(r.Context())

	return &sseStream{
		w:       w,
		flusher: flusher,
		cancel:  cancel,
		done:    ctx.Done(),
	}, nil
}

type sseStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
	cancel  context.CancelFunc
	done    <-chan struct{}
	mu      sync.Mutex
}

// Send writes an event to the SSE stream in standard SSE format.
func (s *sseStream) Send(_ context.Context, event *eventstream.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return errors.New("stream closed")
	default:
	}

	if event.Type != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event.Type); err != nil {
			return fmt.Errorf("writing event type: %w", err)
		}
	}

	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", event.Payload); err != nil {
		return fmt.Errorf("writing event data: %w", err)
	}

	s.flusher.Flush()

	return nil
}

// Done returns a channel that closes when the stream terminates.
func (s *sseStream) Done() <-chan struct{} {
	return s.done
}

// Close terminates the SSE stream.
func (s *sseStream) Close() error {
	s.cancel()
	return nil
}
