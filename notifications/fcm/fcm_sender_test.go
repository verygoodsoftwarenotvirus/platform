package fcm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSender(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	logger := logging.NewNoopLogger()
	tracingProvider := tracing.NewNoopTracerProvider()

	t.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		sender, err := NewSender(ctx, nil, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config is required")
	})

	t.Run("with non-existent credentials path", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			CredentialsPath: filepath.Join(t.TempDir(), "nonexistent-firebase-credentials.json"),
		}
		sender, err := NewSender(ctx, cfg, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credentials file not found")
	})

	t.Run("with empty credentials path uses ADC", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{CredentialsPath: ""}
		sender, err := NewSender(ctx, cfg, tracingProvider, logger)
		// ADC typically fails without GCP credentials in test env
		if err != nil {
			assert.Nil(t, sender)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "fcm:")
			return
		}
		require.NotNil(t, sender)
	})

	t.Run("with invalid JSON credentials file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "creds.json")
		require.NoError(t, os.WriteFile(path, []byte("not valid json"), 0o600))

		cfg := &Config{CredentialsPath: path}
		sender, err := NewSender(ctx, cfg, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fcm:")
	})
}
