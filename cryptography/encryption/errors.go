package encryption

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
)

var (
	ErrIncorrectKeyLength = errors.New("secret is not the right length")
)
