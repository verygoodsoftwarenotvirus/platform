package http

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v5/database"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v5/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v5/types"

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

	T.Run("unknown error returns fallback", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(errors.New("totally unknown error that no mapper handles"))
		assert.Equal(t, types.ErrTalkingToDatabase, code)
		assert.Equal(t, "an error occurred", msg)
	})

	T.Run("circuit broken error", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(circuitbreaking.ErrCircuitBroken)
		assert.Equal(t, types.ErrCircuitBroken, code)
		assert.Equal(t, "service temporarily unavailable", msg)
	})

	T.Run("ErrNilInputParameter", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(platformerrors.ErrNilInputParameter)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
		assert.Equal(t, "invalid input", msg)
	})

	T.Run("ErrUserAlreadyExists", func(t *testing.T) {
		t.Parallel()
		code, msg := ToAPIError(database.ErrUserAlreadyExists)
		assert.Equal(t, types.ErrValidatingRequestInput, code)
		assert.Equal(t, "user already exists", msg)
	})
}

type testHTTPMapper struct {
	err  error
	code types.ErrorCode
	msg  string
}

func (m testHTTPMapper) Map(err error) (types.ErrorCode, string, bool) {
	if errors.Is(err, m.err) {
		return m.code, m.msg, true
	}
	return "", "", false
}

func TestRegisterHTTPErrorMapper(T *testing.T) {
	T.Parallel()

	T.Run("registers a mapper that is consulted by ToAPIError", func(t *testing.T) {
		t.Parallel()

		customErr := errors.New("http-register-test-error")
		mapper := testHTTPMapper{err: customErr, code: "E_CUSTOM", msg: "custom message"}

		RegisterHTTPErrorMapper(mapper)

		code, msg := ToAPIError(customErr)
		assert.Equal(t, types.ErrorCode("E_CUSTOM"), code)
		assert.Equal(t, "custom message", msg)
	})
}
