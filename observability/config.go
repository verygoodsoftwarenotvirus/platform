package observability

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/logging/config"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/platform/observability/profiling"
	profilingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/profiling/config"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/tracing/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config contains settings about how we report our metrics.
	Config struct {
		_         struct{}            `json:"-"`
		Profiling profilingcfg.Config `envPrefix:"PROFILING_" json:"profiling"`
		Logging   loggingcfg.Config   `envPrefix:"LOGGING_"   json:"logging"`
		Metrics   metricscfg.Config   `envPrefix:"METRICS_"   json:"metrics"`
		Tracing   tracingcfg.Config   `envPrefix:"TRACING_"   json:"tracing"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Logging),
		validation.Field(&cfg.Metrics),
		validation.Field(&cfg.Tracing),
		validation.Field(&cfg.Profiling),
	)
}

// Pillars holds the four observability pillars: logging, tracing, metrics, and profiling.
type Pillars struct {
	Logger          logging.Logger
	TracerProvider  tracing.TracerProvider
	MetricsProvider metrics.Provider
	Profiler        profiling.Provider
}

// ProvidePillars creates and returns all four observability pillars.
func (cfg *Config) ProvidePillars(ctx context.Context) (*Pillars, error) {
	logger, err := cfg.Logging.ProvideLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("setting up logger: %w", err)
	}

	tracerProvider, err := cfg.Tracing.ProvideTracerProvider(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("setting up tracer provider: %w", err)
	}

	metricsProvider, err := cfg.Metrics.ProvideMetricsProvider(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("setting up metrics provider: %w", err)
	}

	profiler, err := cfg.Profiling.ProvideProfilingProvider(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("setting up profiling provider: %w", err)
	}

	return &Pillars{
		Logger:          logger,
		TracerProvider:  tracerProvider,
		MetricsProvider: metricsProvider,
		Profiler:        profiler,
	}, nil
}
