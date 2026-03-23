package featureflags

import (
	"context"
)

type (
	// FeatureFlagManager manages feature flags.
	FeatureFlagManager interface {
		CanUseFeature(ctx context.Context, userID, feature string) (bool, error)
		GetStringValue(ctx context.Context, userID, feature string) (string, error)
		GetInt64Value(ctx context.Context, userID, feature string) (int64, error)
		Close() error
	}
)

// NewNoopFeatureFlagManager returns a FeatureFlagManager that always returns zero values.
func NewNoopFeatureFlagManager() FeatureFlagManager {
	return &NoopFeatureFlagManager{}
}

// NoopFeatureFlagManager is a no-op FeatureFlagManager.
type NoopFeatureFlagManager struct{}

// CanUseFeature implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) CanUseFeature(context.Context, string, string) (bool, error) {
	return false, nil
}

// GetStringValue implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) GetStringValue(context.Context, string, string) (string, error) {
	return "", nil
}

// GetInt64Value implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) GetInt64Value(context.Context, string, string) (int64, error) {
	return 0, nil
}

// Close implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) Close() error {
	return nil
}
