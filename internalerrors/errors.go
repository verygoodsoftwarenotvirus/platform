package internalerrors

import (
	"fmt"
)

// NilConfigError returns a nil config error.
func NilConfigError(name string) error {
	return fmt.Errorf("nil config provided for %s", name)
}
