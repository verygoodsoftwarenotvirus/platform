package mockuploads

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/uploads"

	"github.com/stretchr/testify/mock"
)

var _ uploads.UploadManager = (*MockUploadManager)(nil)

// MockUploadManager is a mock MockUploadManager.
type MockUploadManager struct {
	mock.Mock
}

// SaveFile satisfies the MockUploadManager interface.
func (m *MockUploadManager) SaveFile(ctx context.Context, path string, content []byte) error {
	return m.Called(ctx, path, content).Error(0)
}

// ReadFile satisfies the MockUploadManager interface.
func (m *MockUploadManager) ReadFile(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)

	return args.Get(0).([]byte), args.Error(1)
}
