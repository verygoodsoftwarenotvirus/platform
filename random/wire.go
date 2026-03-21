package random

import "github.com/google/wire"

// RandProviders are what we offer to dependency injection.
var RandProviders = wire.NewSet(
	NewGenerator,
)
