package postgres

import (
	"github.com/google/wire"
)

var (
	// PGProviders are what we offer to dependency injection.
	PGProviders = wire.NewSet(
		ProvideDatabaseClient,
	)
)
