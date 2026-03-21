package http

import (
	"github.com/google/wire"
)

// ProvidersHTTP are our wire superset of providers this package offers.
var ProvidersHTTP = wire.NewSet(
	ProvideHTTPServer,
)
