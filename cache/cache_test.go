package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrNotFound(T *testing.T) {
	T.Parallel()

	T.Run("is not nil", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrNotFound)
		assert.Equal(t, "not found", ErrNotFound.Error())
	})
}
