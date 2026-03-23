package types

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

func TestErrorResponse_Error(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, (&APIError{}).Error())
	})
}

func TestAPIResponse_EncodeToJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		example := &APIResponse[string]{
			Error: &APIError{
				Message: t.Name(),
				Code:    ErrDataNotFound,
			},
		}

		encodedBytes, err := json.Marshal(example)
		require.NoError(t, err)

		expected := `{"error":{"message":"TestAPIResponse_EncodeToJSON/standard","code":"E104"},"details":{"currentAccountID":"","traceID":""}}`
		actual := string(encodedBytes)

		assert.Equal(t, expected, actual)
	})
}

func TestAPIError_AsError(T *testing.T) {
	T.Parallel()

	T.Run("with nil receiver", func(t *testing.T) {
		t.Parallel()

		var e *APIError
		assert.NoError(t, e.AsError())
	})

	T.Run("with non-nil receiver", func(t *testing.T) {
		t.Parallel()

		e := &APIError{
			Message: "something went wrong",
			Code:    ErrNothingSpecific,
		}
		assert.Error(t, e.AsError())
	})
}

func TestNewAPIErrorResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		details := ResponseDetails{
			CurrentAccountID: "account123",
			TraceID:          "trace456",
		}

		resp := NewAPIErrorResponse("something broke", ErrTalkingToDatabase, details)

		require.NotNil(t, resp)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "something broke", resp.Error.Message)
		assert.Equal(t, ErrTalkingToDatabase, resp.Error.Code)
		assert.Equal(t, details, resp.Details)
	})
}

func TestFloat32RangeWithOptionalMax_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()

		x := &Float32RangeWithOptionalMax{Min: 1.0}
		assert.NoError(t, x.ValidateWithContext(context.Background()))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		x := &Float32RangeWithOptionalMax{}
		assert.Error(t, x.ValidateWithContext(context.Background()))
	})
}

func TestUint16RangeWithOptionalMax_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()

		x := &Uint16RangeWithOptionalMax{Min: 1}
		assert.NoError(t, x.ValidateWithContext(context.Background()))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		x := &Uint16RangeWithOptionalMax{}
		assert.Error(t, x.ValidateWithContext(context.Background()))
	})
}

func TestUint32RangeWithOptionalMax_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()

		x := &Uint32RangeWithOptionalMax{Min: 1}
		assert.NoError(t, x.ValidateWithContext(context.Background()))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		x := &Uint32RangeWithOptionalMax{}
		assert.Error(t, x.ValidateWithContext(context.Background()))
	})
}

func TestRangeWithOptionalUpperBound_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid", func(t *testing.T) {
		t.Parallel()

		x := &RangeWithOptionalUpperBound[string]{Min: "a"}
		assert.NoError(t, x.ValidateWithContext(context.Background()))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		x := &RangeWithOptionalUpperBound[string]{}
		assert.Error(t, x.ValidateWithContext(context.Background()))
	})
}
