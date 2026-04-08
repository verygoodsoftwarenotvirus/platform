package noop

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/featureflags"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func evalCtx() featureflags.EvaluationContext {
	return featureflags.EvaluationContext{TargetingKey: "user-id"}
}

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

	T.Run("returns false", func(t *testing.T) {
		t.Parallel()

		result, err := NewFeatureFlagManager().CanUseFeature(context.Background(), "some-feature", evalCtx())
		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestFeatureFlagManager_GetStringValue(T *testing.T) {
	T.Parallel()

	T.Run("returns default", func(t *testing.T) {
		t.Parallel()

		result, err := NewFeatureFlagManager().GetStringValue(context.Background(), "some-feature", "fallback", evalCtx())
		assert.NoError(t, err)
		assert.Equal(t, "fallback", result)
	})
}

func TestFeatureFlagManager_GetInt64Value(T *testing.T) {
	T.Parallel()

	T.Run("returns default", func(t *testing.T) {
		t.Parallel()

		result, err := NewFeatureFlagManager().GetInt64Value(context.Background(), "some-feature", int64(42), evalCtx())
		assert.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})
}

func TestFeatureFlagManager_GetFloat64Value(T *testing.T) {
	T.Parallel()

	T.Run("returns default", func(t *testing.T) {
		t.Parallel()

		result, err := NewFeatureFlagManager().GetFloat64Value(context.Background(), "some-feature", 3.14, evalCtx())
		assert.NoError(t, err)
		assert.InDelta(t, 3.14, result, 1e-9)
	})
}

func TestFeatureFlagManager_GetObjectValue(T *testing.T) {
	T.Parallel()

	T.Run("returns default", func(t *testing.T) {
		t.Parallel()

		def := map[string]any{"k": "v"}
		result, err := NewFeatureFlagManager().GetObjectValue(context.Background(), "some-feature", def, evalCtx())
		assert.NoError(t, err)
		assert.Equal(t, def, result)
	})
}

func TestFeatureFlagManager_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		err := NewFeatureFlagManager().Close()
		assert.NoError(t, err)
	})
}
