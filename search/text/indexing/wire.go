package indexing

import (
	"github.com/google/wire"
)

var (
	ProvidersIndexing = wire.NewSet(
		NewIndexScheduler,
	)
)
