package reflection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type exampleStruct struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func TestGetTagNameByFieldName(T *testing.T) {
	T.Parallel()

	T.Run("with pointer", func(t *testing.T) {
		t.Parallel()

		x := &exampleStruct{}
		expected := "field1"

		actual, err := GetTagNameByValue(x, x.Field1, "json")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("unpointered", func(t *testing.T) {
		t.Parallel()

		x := exampleStruct{}
		expected := "field1"

		actual, err := GetTagNameByValue(x, x.Field1, "json")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
