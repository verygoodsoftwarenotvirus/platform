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
