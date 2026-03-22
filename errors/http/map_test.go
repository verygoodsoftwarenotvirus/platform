package http

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v2/database"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/types"

	"github.com/stretchr/testify/assert"
)

func TestPlatformMapper_Map(T *testing.T) {
	T.Parallel()

	T.Run("nil error returns ok=false", func(t *testing.T) {
		t.Parallel()
		_, _, ok := PlatformMapper.Map(nil)
		assert.False(t, ok)
	})

	T.Run("sql.ErrNoRows maps to ErrDataNotFound", func(t *testing.T) {
		t.Parallel()
		code, msg, ok := PlatformMapper.Map(sql.ErrNoRows)
		assert.True(t, ok)
		assert.Equal(t, types.ErrDataNotFound, code)
		assert.Equal(t, "data not found", msg)
	})

	T.Run("ErrUserAlreadyExists maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, msg, ok := PlatformMapper.Map(database.ErrUserAlreadyExists)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
		assert.Equal(t, "user already exists", msg)
	})

	T.Run("ErrCircuitBroken maps to ErrCircuitBroken", func(t *testing.T) {
		t.Parallel()
		code, msg, ok := PlatformMapper.Map(circuitbreaking.ErrCircuitBroken)
		assert.True(t, ok)
		assert.Equal(t, types.ErrCircuitBroken, code)
		assert.Equal(t, "service temporarily unavailable", msg)
	})

	T.Run("ErrNilInputParameter maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, _, ok := PlatformMapper.Map(platformerrors.ErrNilInputParameter)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
	})

	T.Run("ErrEmptyInputParameter maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, _, ok := PlatformMapper.Map(platformerrors.ErrEmptyInputParameter)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
	})

	T.Run("ErrNilInputProvided maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, _, ok := PlatformMapper.Map(platformerrors.ErrNilInputProvided)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
	})

	T.Run("ErrInvalidIDProvided maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, _, ok := PlatformMapper.Map(platformerrors.ErrInvalidIDProvided)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
	})

	T.Run("ErrEmptyInputProvided maps to ErrValidatingRequestInput", func(t *testing.T) {
		t.Parallel()
		code, _, ok := PlatformMapper.Map(platformerrors.ErrEmptyInputProvided)
		assert.True(t, ok)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
	})

	T.Run("unknown error returns ok=false", func(t *testing.T) {
		t.Parallel()
		_, _, ok := PlatformMapper.Map(errors.New("nope"))
		assert.False(t, ok)
	})
}

func TestToAPIError(T *testing.T) {
	T.Parallel()

	T.Run("nil error", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(nil)
		assert.Equal(t, types.ErrNothingSpecific, code)
		assert.Empty(t, msg)
	})

	T.Run("known platform error uses PlatformMapper", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(sql.ErrNoRows)
		assert.Equal(t, types.ErrDataNotFound, code)
		assert.Equal(t, "data not found", msg)
	})
}
