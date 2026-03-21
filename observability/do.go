package observability

import (
	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/observability/metrics/config"
	profilingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/profiling/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/observability/tracing/config"

	"github.com/samber/do/v2"
)

// RegisterO11yConfigs registers sub-configs extracted from *Config with the injector.
// This mirrors the wire.FieldsOf pattern in wire.go.
// Prerequisite: *Config must be registered in the injector before calling this.
func RegisterO11yConfigs(i do.Injector) {
	do.Provide[*loggingcfg.Config](i, func(i do.Injector) (*loggingcfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Logging, nil
	})
	do.Provide[*metricscfg.Config](i, func(i do.Injector) (*metricscfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Metrics, nil
	})
	do.Provide[*tracingcfg.Config](i, func(i do.Injector) (*tracingcfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Tracing, nil
	})
	do.Provide[*profilingcfg.Config](i, func(i do.Injector) (*profilingcfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Profiling, nil
	})
}
