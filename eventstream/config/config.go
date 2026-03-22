package config

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v2/eventstream"
	"github.com/verygoodsoftwarenotvirus/platform/v2/eventstream/sse"
	"github.com/verygoodsoftwarenotvirus/platform/v2/eventstream/websocket"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderSSE is the SSE provider.
	ProviderSSE = "sse"
	// ProviderWebSocket is the websocket provider.
	ProviderWebSocket = "websocket"
)

type (
	// Config is the configuration for the event stream provider.
	Config struct {
		WebSocket *websocket.Config `env:"init"     envPrefix:"WEBSOCKET_" json:"websocket,omitempty"`
		Provider  string            `env:"PROVIDER" json:"provider"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ProviderSSE, ProviderWebSocket)),
		validation.Field(&cfg.WebSocket, validation.When(cfg.Provider == ProviderWebSocket, validation.Required)),
	)
}

// ProvideEventStreamUpgrader provides an EventStreamUpgrader based on configuration.
func ProvideEventStreamUpgrader(tracerProvider tracing.TracerProvider, cfg *Config) (eventstream.EventStreamUpgrader, error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case ProviderSSE:
		return sse.NewUpgrader(tracerProvider), nil
	case ProviderWebSocket:
		return websocket.NewUpgrader(tracerProvider, cfg.WebSocket), nil
	default:
		return nil, fmt.Errorf("invalid eventstream provider: %q", cfg.Provider)
	}
}

// ProvideBidirectionalEventStreamUpgrader provides a BidirectionalEventStreamUpgrader based on configuration.
func ProvideBidirectionalEventStreamUpgrader(tracerProvider tracing.TracerProvider, cfg *Config) (eventstream.BidirectionalEventStreamUpgrader, error) {
	switch strings.TrimSpace(strings.ToLower(cfg.Provider)) {
	case ProviderSSE:
		return nil, errors.New("SSE does not support bidirectional event streams")
	case ProviderWebSocket:
		return websocket.NewUpgrader(tracerProvider, cfg.WebSocket), nil
	default:
		return nil, fmt.Errorf("invalid eventstream provider: %q", cfg.Provider)
	}
}
