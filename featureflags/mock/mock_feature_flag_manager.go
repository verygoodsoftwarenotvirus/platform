package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/featureflags"

	"github.com/stretchr/testify/mock"
)

var _ featureflags.FeatureFlagManager = (*FeatureFlagManager)(nil)

type FeatureFlagManager struct {
	mock.Mock
}

// CanUseFeature satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) CanUseFeature(ctx context.Context, username, feature string) (bool, error) {
	returnValues := m.Called(ctx, username, feature)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetStringValue satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetStringValue(ctx context.Context, userID, feature string) (string, error) {
	returnValues := m.Called(ctx, userID, feature)
	return returnValues.String(0), returnValues.Error(1)
}

// GetInt64Value satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetInt64Value(ctx context.Context, userID, feature string) (int64, error) {
	returnValues := m.Called(ctx, userID, feature)
	return returnValues.Get(0).(int64), returnValues.Error(1)
}

// Close satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) Close() error {
	return m.Called().Error(0)
}
