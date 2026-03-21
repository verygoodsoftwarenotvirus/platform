package adler32

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_adler32Hasher_Hash(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		hasher := NewAdler32Hasher()

		result, err := hasher.Hash(t.Name())
		assert.NoError(t, err)
		assert.Equal(t, "546573745f61646c657233324861736865725f486173682f7374616e6461726400000001", result)
	})
}
