package images

import (
	"github.com/google/wire"
)

var (
	// ProvidersImages represents what this library offers to external users in the form of dependencies.
	ProvidersImages = wire.NewSet(
		NewMediaUploadProcessor,
	)
)
