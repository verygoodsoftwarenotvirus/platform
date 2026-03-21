package tracing

import "github.com/google/wire"

var (
	// ProvidersTracing provided HTTP client construction; use httpclient.Providers instead.
	ProvidersTracing = wire.NewSet()
)
