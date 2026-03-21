package encryption

import (
	"errors"
)

var (
	ErrIncorrectKeyLength = errors.New("secret is not the right length")
)
