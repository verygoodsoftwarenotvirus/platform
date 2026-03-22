package config

import (
	"github.com/google/wire"
)

var (
	// Providers provides event stream construction for dependency injection.
	Providers = wire.NewSet(
		ProvideEventStreamUpgrader,
	)

	// BidirectionalProviders provides bidirectional event stream construction for dependency injection.
	BidirectionalProviders = wire.NewSet(
		ProvideBidirectionalEventStreamUpgrader,
	)
)
