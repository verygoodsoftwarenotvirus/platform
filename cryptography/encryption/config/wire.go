package config

import (
	"github.com/google/wire"
)

var (
	// Providers provides encryption construction for dependency injection.
	Providers = wire.NewSet(
		ProvideEncryptorDecryptor,
	)
)
