package httpclient

import (
	"net/http"

	"github.com/google/wire"
)

var (
	// Providers provides HTTP client construction for dependency injection.
	Providers = wire.NewSet(
		ProvideHTTPClient,
	)
)

// ProvideHTTPClient provides an HTTP client from config.
// If cfg is nil, defaults are used.
func ProvideHTTPClient(cfg *Config) *http.Client {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.EnsureDefaults()
	return cfg.BuildClient()
}
