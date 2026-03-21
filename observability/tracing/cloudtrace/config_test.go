package cloudtrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloudTraceConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			ProjectID: t.Name(),
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}
