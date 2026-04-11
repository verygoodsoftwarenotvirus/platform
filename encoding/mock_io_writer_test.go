package encoding

import (
	"io"
)

var _ io.Writer = (*mockWriter)(nil)

// mockWriter mocks an io.Writer.
type mockWriter struct {
	WriteFunc func(p []byte) (int, error)
}

// Write implements the io.Writer interface.
func (m *mockWriter) Write(p []byte) (int, error) {
	return m.WriteFunc(p)
}
