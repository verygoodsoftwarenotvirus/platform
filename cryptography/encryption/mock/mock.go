package encryptionmock

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockImpl is the mock EncryptorDecryptor implementation.
type MockImpl struct {
	mock.Mock
}

func NewMockEncryptorDecryptor() *MockImpl {
	return &MockImpl{}
}

// Encrypt is a mock method.
func (m *MockImpl) Encrypt(ctx context.Context, content string) (string, error) {
	returnVals := m.Called(ctx, content)
	return returnVals.String(0), returnVals.Error(1)
}

// Decrypt is a mock method.
func (m *MockImpl) Decrypt(ctx context.Context, content string) (string, error) {
	returnVals := m.Called(ctx, content)
	return returnVals.String(0), returnVals.Error(1)
}
