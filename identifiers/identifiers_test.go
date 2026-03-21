package identifiers

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestNew(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := New()
		assert.NotEmpty(t, actual)
	})
}

func TestValidate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := Validate(xid.New().String())
		assert.NoError(t, actual)
	})
}
