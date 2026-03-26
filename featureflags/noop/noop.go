package noop

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/featureflags"
)

var _ featureflags.FeatureFlagManager = (*featureFlagManager)(nil)

// featureFlagManager is a no-op FeatureFlagManager.
type featureFlagManager struct{}

// NewFeatureFlagManager returns a FeatureFlagManager that always returns zero values.
func NewFeatureFlagManager() featureflags.FeatureFlagManager {
	return &featureFlagManager{}
}

// CanUseFeature implements the FeatureFlagManager interface.
func (*featureFlagManager) CanUseFeature(context.Context, string, string) (bool, error) {
	return false, nil
}

// GetStringValue implements the FeatureFlagManager interface.
func (*featureFlagManager) GetStringValue(context.Context, string, string) (string, error) {
	return "", nil
}

// GetInt64Value implements the FeatureFlagManager interface.
func (*featureFlagManager) GetInt64Value(context.Context, string, string) (int64, error) {
	return 0, nil
}

// Close implements the FeatureFlagManager interface.
func (*featureFlagManager) Close() error {
	return nil
}
