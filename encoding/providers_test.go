package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvideContentType(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeJSON, ProvideContentType(Config{ContentType: "application/json"}))
	})
}
