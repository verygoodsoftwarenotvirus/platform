package observability

import (
	"github.com/google/wire"
)

var (
	O11yProviders = wire.NewSet(
		wire.FieldsOf(
			new(*Config),
			"Logging",
			"Metrics",
			"Tracing",
			"Profiling",
		),
	)
)
