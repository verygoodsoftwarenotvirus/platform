package types

import (
	"encoding/json"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
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
