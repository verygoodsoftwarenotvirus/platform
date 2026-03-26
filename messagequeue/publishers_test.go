package messagequeue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrEmptyTopicName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ErrEmptyTopicName)
		assert.Error(t, ErrEmptyTopicName)
	})
}
