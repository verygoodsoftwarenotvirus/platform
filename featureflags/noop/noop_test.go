package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFeatureFlagManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewFeatureFlagManager()
		require.NotNil(t, mgr)
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewFeatureFlagManager()

		result, err := mgr.CanUseFeature(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestFeatureFlagManager_GetStringValue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewFeatureFlagManager()

		result, err := mgr.GetStringValue(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestFeatureFlagManager_GetInt64Value(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewFeatureFlagManager()

		result, err := mgr.GetInt64Value(context.Background(), "user-id", "some-feature")
		assert.NoError(t, err)
		assert.Zero(t, result)
	})
}

func TestFeatureFlagManager_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mgr := NewFeatureFlagManager()

		err := mgr.Close()
		assert.NoError(t, err)
	})
}
