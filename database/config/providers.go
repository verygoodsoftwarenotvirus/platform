package databasecfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
)

// ProvideClientConfig converts Config to database.ClientConfig.
//
//nolint:gocritic // hugeParam: intentionally accepts value for compatibility
func ProvideClientConfig(cfg Config) database.ClientConfig {
	return &cfg
}
