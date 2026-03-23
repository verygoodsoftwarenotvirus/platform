package mocksearch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIndexManager_Index(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &IndexManager[string]{}
		m.On("Index", mock.Anything, "id", "value").Return(nil)

		err := m.Index(context.Background(), "id", "value")
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, m)
	})
}

func TestIndexManager_Search(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := []*string{new(string), new(string)}
		*expected[0] = "result1"
		*expected[1] = "result2"

		m := &IndexManager[string]{}
		m.On("Search", mock.Anything, "query").Return(expected, nil)

		results, err := m.Search(context.Background(), "query")
		assert.NoError(t, err)
		assert.Equal(t, expected, results)

		mock.AssertExpectationsForObjects(t, m)
	})
}

func TestIndexManager_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &IndexManager[string]{}
		m.On("Delete", mock.Anything, "id").Return(nil)

		err := m.Delete(context.Background(), "id")
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, m)
	})
}

func TestIndexManager_Wipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		m := &IndexManager[string]{}
		m.On("Wipe", mock.Anything).Return(nil)

		err := m.Wipe(context.Background())
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, m)
	})
}
