package testutils

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.Writer = (*MockWriter)(nil)

// MockWriter mocks a io.Writer.
type MockWriter struct {
	mock.Mock
}

// Write implements the io.Writer interface.
func (m *MockWriter) Write(p []byte) (int, error) {
	returnVals := m.Called(p)
	return returnVals.Int(0), returnVals.Error(1)
}
