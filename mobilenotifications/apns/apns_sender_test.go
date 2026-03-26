package apns

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestP8File(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "AuthKey.p8")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))
	return path
}

func TestNewSender(T *testing.T) {
	T.Parallel()

	logger := logging.NewNoopLogger()
	tracingProvider := tracing.NewNoopTracerProvider()

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		sender, err := NewSender(nil, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required config")
	})

	T.Run("with empty auth key path", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			KeyID:      "KEY123",
			TeamID:     "TEAM123",
			BundleID:   "com.example.app",
			Production: false,
		}
		sender, err := NewSender(cfg, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required config")
	})

	T.Run("with empty key ID", func(t *testing.T) {
		t.Parallel()

		p8Path := createTestP8File(t)
		cfg := &Config{
			AuthKeyPath: p8Path,
			TeamID:      "TEAM123",
			BundleID:    "com.example.app",
			Production:  false,
		}
		sender, err := NewSender(cfg, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required config")
	})

	T.Run("with non-existent auth key file", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			AuthKeyPath: filepath.Join(t.TempDir(), "nonexistent.p8"),
			KeyID:       "KEY123",
			TeamID:      "TEAM123",
			BundleID:    "com.example.app",
			Production:  false,
		}
		sender, err := NewSender(cfg, tracingProvider, logger)
		assert.Nil(t, sender)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loading auth key")
	})

	T.Run("with valid config", func(t *testing.T) {
		t.Parallel()

		p8Path := createTestP8File(t)
		cfg := &Config{
			AuthKeyPath: p8Path,
			KeyID:       "KEY123",
			TeamID:      "TEAM123",
			BundleID:    "com.example.app",
			Production:  false,
		}
		sender, err := NewSender(cfg, tracingProvider, logger)
		require.NoError(t, err)
		require.NotNil(t, sender)
		assert.Equal(t, "com.example.app", sender.topic)
	})
}

func TestSender_Send_rejectsInvalidDeviceToken(T *testing.T) {
	T.Parallel()

	p8Path := createTestP8File(T)
	cfg := &Config{
		AuthKeyPath: p8Path,
		KeyID:       "KEY123",
		TeamID:      "TEAM123",
		BundleID:    "com.example.app",
		Production:  false,
	}
	sender, err := NewSender(cfg, tracing.NewNoopTracerProvider(), logging.NewNoopLogger())
	require.NoError(T, err)

	ctx := T.Context()

	T.Run("rejects binary/garbage token", func(t *testing.T) {
		t.Parallel()
		// Simulates decrypted garbage (e.g. wrong key or corrupted data)
		invalidToken := "x\x89\xbf\x1f\xa0\x93\x12\xf5"
		sendErr := sender.Send(ctx, invalidToken, "title", "body", nil)
		require.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "invalid device token format")
	})

	T.Run("rejects token with control characters", func(t *testing.T) {
		t.Parallel()
		invalidToken := "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef12345\t"
		sendErr := sender.Send(ctx, invalidToken, "title", "body", nil)
		require.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "invalid device token format")
	})

	T.Run("rejects too short token", func(t *testing.T) {
		t.Parallel()
		sendErr := sender.Send(ctx, "abc123", "title", "body", nil)
		require.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "invalid device token format")
	})
}
