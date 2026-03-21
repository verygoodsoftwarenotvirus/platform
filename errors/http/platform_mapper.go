package http

import (
	"database/sql"
	"errors"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/database"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/errors"
	"github.com/verygoodsoftwarenotvirus/platform/types"
)

// PlatformMapper maps platform-level errors to HTTP error codes and messages.
// It does not depend on any domain (mealplanning, etc.).
var PlatformMapper HTTPErrorMapper = platformMapper{}

type platformMapper struct{}

func (platformMapper) Map(err error) (code types.ErrorCode, msg string, ok bool) {
	if err == nil {
		return types.ErrNothingSpecific, "", false
	}
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return types.ErrDataNotFound, "data not found", true
	case errors.Is(err, database.ErrUserAlreadyExists):
		return types.ErrValidatingRequestInput, "user already exists", true
	case errors.Is(err, circuitbreaking.ErrCircuitBroken):
		return types.ErrCircuitBroken, "service temporarily unavailable", true
	case errors.Is(err, platformerrors.ErrNilInputParameter),
		errors.Is(err, platformerrors.ErrEmptyInputParameter),
		errors.Is(err, platformerrors.ErrNilInputProvided),
		errors.Is(err, platformerrors.ErrInvalidIDProvided),
		errors.Is(err, platformerrors.ErrEmptyInputProvided):
		return types.ErrValidatingRequestInput, "invalid input", true
	default:
		return "", "", false
	}
}
