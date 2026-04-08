package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/uploads/objectstorage"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterStorageConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue(i, &Config{
			Storage: objectstorage.Config{
				BucketName: t.Name(),
				Provider:   objectstorage.MemoryProvider,
			},
		})

		RegisterStorageConfig(i)

		storageCfg, err := do.Invoke[*objectstorage.Config](i)
		require.NoError(t, err)
		assert.NotNil(t, storageCfg)
		assert.Equal(t, t.Name(), storageCfg.BucketName)
	})
}
