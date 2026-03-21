package testutils

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.ReadCloser = (*MockReadCloser)(nil)

// MockReadCloser mocks a io.ReadCloser.
type MockReadCloser struct {
	mock.Mock
}

// Read implements the io.ReadCloser interface.
func (m *MockReadCloser) Read(p []byte) (int, error) {
	returnValues := m.Called(p)

	return returnValues.Int(0), returnValues.Error(1)
}

// Close implements the io.ReadCloser interface.
func (m *MockReadCloser) Close() error {
	return m.Called().Error(0)
}
