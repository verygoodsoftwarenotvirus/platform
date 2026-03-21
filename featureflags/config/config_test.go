package featureflagscfg

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/featureflags/launchdarkly"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			LaunchDarkly: &launchdarkly.Config{
				SDKKey:      t.Name(),
				InitTimeout: 123,
			},
			Provider: ProviderLaunchDarkly,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty provider for noop", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "",
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
