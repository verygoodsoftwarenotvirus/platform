package objectstorage

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/uploads"
)

// ProvideUploadManager transforms an *objectstorage.Uploader into an UploadManager.
func ProvideUploadManager(u *Uploader) uploads.UploadManager {
	return u
}
