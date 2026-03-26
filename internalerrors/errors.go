package internalerrors

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
)

// NilConfigError returns a nil config error.
func NilConfigError(name string) error {
	return errors.Newf("nil config provided for %s", name)
}
