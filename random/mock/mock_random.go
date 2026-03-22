package randommock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/random"

	"github.com/stretchr/testify/mock"
)

var _ random.Generator = (*Generator)(nil)

// Generator is a mock Generator.
type Generator struct {
	mock.Mock
}

func (m *Generator) GenerateHexEncodedString(ctx context.Context, length int) (string, error) {
	returnVals := m.Called(ctx, length)
	return returnVals.String(0), returnVals.Error(1)
}

// GenerateBase32EncodedString implements our interface.
func (m *Generator) GenerateBase32EncodedString(ctx context.Context, length int) (string, error) {
	returnVals := m.Called(ctx, length)

	return returnVals.String(0), returnVals.Error(1)
}

// GenerateBase64EncodedString implements our interface.
func (m *Generator) GenerateBase64EncodedString(ctx context.Context, length int) (string, error) {
	returnVals := m.Called(ctx, length)

	return returnVals.String(0), returnVals.Error(1)
}

// GenerateRawBytes implements our interface.
func (m *Generator) GenerateRawBytes(ctx context.Context, length int) ([]byte, error) {
	returnVals := m.Called(ctx, length)

	return returnVals.Get(0).([]byte), returnVals.Error(1)
}
