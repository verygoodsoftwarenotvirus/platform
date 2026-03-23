package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrIncorrectKeyLength(T *testing.T) {
	T.Parallel()

	T.Run("is not nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrIncorrectKeyLength)
		assert.Equal(t, "secret is not the right length", ErrIncorrectKeyLength.Error())
	})
}
