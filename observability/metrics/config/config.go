package metricscfg

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics/otelgrpc"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderOtel represents the open source tracing server.
	ProviderOtel = "otelgrpc"
)

type (
	// Config contains settings related to tracing.
	Config struct {
		_ struct{} `json:"-"`

		Otel        *otelgrpc.Config `env:"init"         envPrefix:"OTEL_"         json:"otelgrpc,omitempty"`
		ServiceName string           `env:"SERVICE_NAME" json:"serviceName"`
		Provider    string           `env:"PROVIDER"     json:"provider,omitempty"`
	}
)

// ProvideMetricsProvider provides a metrics provider.
func (c *Config) ProvideMetricsProvider(ctx context.Context, logger logging.Logger) (metrics.Provider, error) {
	p := strings.TrimSpace(strings.ToLower(c.Provider))

	switch p {
	case ProviderOtel:
		return otelgrpc.ProvideMetricsProvider(ctx, logger, c.ServiceName, c.Otel)
	default:
		return metrics.NewNoopMetricsProvider(), nil
	}
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the config struct.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Provider, validation.In("", ProviderOtel)),
		validation.Field(&c.Otel, validation.When(c.Provider == ProviderOtel, validation.Required).Else(validation.Nil)),
	)
}
