package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"

	"github.com/stretchr/testify/mock"
)

var _ featureflags.FeatureFlagManager = (*FeatureFlagManager)(nil)

type FeatureFlagManager struct {
	mock.Mock
}

// CanUseFeature satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) CanUseFeature(ctx context.Context, feature string, evalCtx featureflags.EvaluationContext) (bool, error) {
	returnValues := m.Called(ctx, feature, evalCtx)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetStringValue satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetStringValue(ctx context.Context, feature, defaultValue string, evalCtx featureflags.EvaluationContext) (string, error) {
	returnValues := m.Called(ctx, feature, defaultValue, evalCtx)
	return returnValues.String(0), returnValues.Error(1)
}

// GetInt64Value satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetInt64Value(ctx context.Context, feature string, defaultValue int64, evalCtx featureflags.EvaluationContext) (int64, error) {
	returnValues := m.Called(ctx, feature, defaultValue, evalCtx)
	return returnValues.Get(0).(int64), returnValues.Error(1)
}

// GetFloat64Value satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetFloat64Value(ctx context.Context, feature string, defaultValue float64, evalCtx featureflags.EvaluationContext) (float64, error) {
	returnValues := m.Called(ctx, feature, defaultValue, evalCtx)
	return returnValues.Get(0).(float64), returnValues.Error(1)
}

// GetObjectValue satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) GetObjectValue(ctx context.Context, feature string, defaultValue any, evalCtx featureflags.EvaluationContext) (any, error) {
	returnValues := m.Called(ctx, feature, defaultValue, evalCtx)
	return returnValues.Get(0), returnValues.Error(1)
}

// Close satisfies the FeatureFlagManager interface.
func (m *FeatureFlagManager) Close() error {
	return m.Called().Error(0)
}
