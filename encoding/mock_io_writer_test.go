package encoding

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.Writer = (*mockWriter)(nil)

// mockWriter mocks a io.Writer.
type mockWriter struct {
	mock.Mock
}

// Write implements the io.Writer interface.
func (m *mockWriter) Write(p []byte) (int, error) {
	returnVals := m.Called(p)
	return returnVals.Int(0), returnVals.Error(1)
}
