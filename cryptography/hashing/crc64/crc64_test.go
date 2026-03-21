package crc64

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_crc64Hasher_Hash(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		hasher := NewCRC64Hasher()

		result, err := hasher.Hash(t.Name())
		assert.NoError(t, err)
		assert.Equal(t, "546573745f63726336344861736865725f486173682f7374616e646172640000000000000000", result)
	})
}
