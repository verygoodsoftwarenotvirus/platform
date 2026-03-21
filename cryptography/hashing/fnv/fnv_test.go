package fnv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fnvHasher_Hash(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		hasher := NewFNVHasher()

		result, err := hasher.Hash(t.Name())
		assert.NoError(t, err)
		assert.Equal(t, "546573745f666e764861736865725f486173682f7374616e646172646c62272e07bb014262b821756295c58d", result)
	})
}
