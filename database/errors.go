package database

import (
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
)

// ErrUserAlreadyExists indicates that a user with that username has already been created.
var ErrUserAlreadyExists = platformerrors.New("user already exists")
