package featureflags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoopFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewNoopFeatureFlagManager()
		require.NotNil(t, mgr)
	})
}

func TestNoopFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewNoopFeatureFlagManager()

		result, err := mgr.CanUseFeature(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestNoopFeatureFlagManager_GetStringValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewNoopFeatureFlagManager()

		result, err := mgr.GetStringValue(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestNoopFeatureFlagManager_GetInt64Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewNoopFeatureFlagManager()

		result, err := mgr.GetInt64Value(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.Zero(t, result)
	})
}

func TestNoopFeatureFlagManager_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewNoopFeatureFlagManager()

		err := mgr.Close()
		assert.NoError(t, err)
	})
}
