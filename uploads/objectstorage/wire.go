package objectstorage

import (
	"github.com/verygoodsoftwarenotvirus/platform/uploads"

	"github.com/google/wire"
)

var (
	// Providers are what we provide to the dependency injection framework.
	Providers = wire.NewSet(
		NewUploadManager,
		ProvideUploadManager,
	)
)

// ProvideUploadManager transforms an *objectstorage.Uploader into an UploadManager.
func ProvideUploadManager(u *Uploader) uploads.UploadManager {
	return u
}
