package objectstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockUploader_SaveFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m := &MockUploader{}
		m.On("SaveFile", mock.Anything, "test.txt", []byte("content")).Return(nil)

		assert.NoError(t, m.SaveFile(ctx, "test.txt", []byte("content")))
		mock.AssertExpectationsForObjects(t, m)
	})
}

func TestMockUploader_ReadFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := []byte("content")
		m := &MockUploader{}
		m.On("ReadFile", mock.Anything, "test.txt").Return(expected, nil)

		actual, err := m.ReadFile(ctx, "test.txt")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
		mock.AssertExpectationsForObjects(t, m)
	})
}
