package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/eventstream"

	gorillawebsocket "github.com/gorilla/websocket"
)

var (
	_ eventstream.EventStreamUpgrader              = (*Upgrader)(nil)
	_ eventstream.BidirectionalEventStreamUpgrader = (*Upgrader)(nil)
	_ eventstream.EventStream                      = (*wsStream)(nil)
	_ eventstream.BidirectionalEventStream         = (*bidirectionalWSStream)(nil)
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultBufferSize        = 1024
	incomingChannelBuffer    = 64
)

// Upgrader upgrades HTTP connections to WebSocket event streams.
type Upgrader struct {
	wsUpgrader        gorillawebsocket.Upgrader
	heartbeatInterval time.Duration
}

// NewUpgrader creates a new WebSocket Upgrader.
func NewUpgrader(cfg *Config) *Upgrader {
	heartbeat := defaultHeartbeatInterval
	readBuf := defaultBufferSize
	writeBuf := defaultBufferSize

	if cfg != nil {
		if cfg.HeartbeatInterval > 0 {
			heartbeat = cfg.HeartbeatInterval
		}
		if cfg.ReadBufferSize > 0 {
			readBuf = cfg.ReadBufferSize
		}
		if cfg.WriteBufferSize > 0 {
			writeBuf = cfg.WriteBufferSize
		}
	}

	return &Upgrader{
		wsUpgrader: gorillawebsocket.Upgrader{
			ReadBufferSize:  readBuf,
			WriteBufferSize: writeBuf,
			CheckOrigin:     func(*http.Request) bool { return true },
		},
		heartbeatInterval: heartbeat,
	}
}

// UpgradeToEventStream upgrades an HTTP connection to a unidirectional WebSocket event stream.
func (u *Upgrader) UpgradeToEventStream(w http.ResponseWriter, r *http.Request) (eventstream.EventStream, error) {
	conn, err := u.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("upgrading to websocket: %w", err)
	}

	return newWSStream(conn, u.heartbeatInterval), nil
}

// UpgradeToBidirectionalStream upgrades an HTTP connection to a bidirectional WebSocket event stream.
func (u *Upgrader) UpgradeToBidirectionalStream(w http.ResponseWriter, r *http.Request) (eventstream.BidirectionalEventStream, error) {
	conn, err := u.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("upgrading to websocket: %w", err)
	}

	return newBidirectionalWSStream(conn, u.heartbeatInterval), nil
}

// wsStream is a unidirectional (send-only) WebSocket event stream.
type wsStream struct {
	conn              *gorillawebsocket.Conn
	heartbeatInterval *time.Ticker
	done              chan struct{}
	mu                sync.Mutex
	closed            bool
}

func newWSStream(conn *gorillawebsocket.Conn, heartbeat time.Duration) *wsStream {
	s := &wsStream{
		conn: conn,
		done: make(chan struct{}),
	}

	if heartbeat > 0 {
		s.heartbeatInterval = time.NewTicker(heartbeat)
		go s.heartbeatLoop()
	}

	return s
}

func (s *wsStream) heartbeatLoop() {
	for {
		select {
		case <-s.done:
			return
		case <-s.heartbeatInterval.C:
			s.mu.Lock()
			err := s.conn.WriteMessage(gorillawebsocket.PingMessage, nil)
			s.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// Send writes an event to the WebSocket connection as JSON.
func (s *wsStream) Send(_ context.Context, event *eventstream.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors.New("stream closed")
	}

	return s.conn.WriteJSON(event)
}

// Done returns a channel that closes when the stream terminates.
func (s *wsStream) Done() <-chan struct{} {
	return s.done
}

// Close terminates the WebSocket stream.
func (s *wsStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	close(s.done)

	if s.heartbeatInterval != nil {
		s.heartbeatInterval.Stop()
	}

	return s.conn.Close()
}

// bidirectionalWSStream extends wsStream with receive capability.
type bidirectionalWSStream struct {
	*wsStream
	incoming chan *eventstream.Event
}

func newBidirectionalWSStream(conn *gorillawebsocket.Conn, heartbeat time.Duration) *bidirectionalWSStream {
	s := &bidirectionalWSStream{
		wsStream: newWSStream(conn, heartbeat),
		incoming: make(chan *eventstream.Event, incomingChannelBuffer),
	}

	go s.readLoop()

	return s
}

func (s *bidirectionalWSStream) readLoop() {
	defer close(s.incoming)

	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			return
		}

		var event eventstream.Event
		if err = json.Unmarshal(msg, &event); err != nil {
			continue
		}

		select {
		case s.incoming <- &event:
		case <-s.done:
			return
		}
	}
}

// Receive returns a channel of inbound events from the client.
func (s *bidirectionalWSStream) Receive() <-chan *eventstream.Event {
	return s.incoming
}
