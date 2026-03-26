package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/mobilenotifications"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"

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

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	ctx := T.Context()

	T.Run("with noop provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderNoop}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with empty provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ""}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with apns_fcm provider and both nil", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderAPNsFCM,
			APNs:     nil,
			FCM:      nil,
		}
		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with apns_fcm provider and nil APNs but FCM present", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderAPNsFCM,
			APNs:     nil,
			FCM:      &FCMConfig{},
		}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with apns_fcm provider and nil FCM but APNs present", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderAPNsFCM,
			APNs:     &APNsConfig{AuthKeyPath: "x", KeyID: "x", TeamID: "x", BundleID: "x"},
			FCM:      nil,
		}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with apns_fcm provider and both configs", func(t *testing.T) {
		t.Parallel()

		p8Path := createTestP8File(t)
		cfg := &Config{
			Provider: ProviderAPNsFCM,
			APNs:     &APNsConfig{AuthKeyPath: p8Path, KeyID: "x", TeamID: "x", BundleID: "x"},
			FCM:      &FCMConfig{},
		}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvidePushSender(T *testing.T) {
	T.Parallel()

	ctx := T.Context()
	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracerProvider()

	T.Run("with empty provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := Config{Provider: ""}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		// Noop returns nil on SendPush
		assert.NoError(t, sender.SendPush(ctx, "ios", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"}))
	})

	T.Run("with noop provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := Config{Provider: ProviderNoop}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		assert.NoError(t, sender.SendPush(ctx, "android", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"}))
	})

	T.Run("with apns_fcm provider and nil APNs returns noop or FCM-only sender", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Provider: ProviderAPNsFCM,
			APNs:     nil,
			FCM:      &FCMConfig{},
		}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		// FCM init typically fails in unit tests (no ADC), so we get noop; if it succeeds, iOS would error
		_ = sender.SendPush(ctx, "ios", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"})
	})

	T.Run("with apns_fcm provider and nil FCM returns iOS-only sender", func(t *testing.T) {
		t.Parallel()

		p8Path := createTestP8File(t)
		cfg := Config{
			Provider: ProviderAPNsFCM,
			APNs:     &APNsConfig{AuthKeyPath: p8Path, KeyID: "x", TeamID: "x", BundleID: "x"},
			FCM:      nil,
		}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		// Android not configured, should return ErrPlatformNotSupported
		err = sender.SendPush(ctx, "android", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, mobilenotifications.ErrPlatformNotSupported)
	})

	T.Run("with apns_fcm provider and invalid APNs path returns noop or FCM-only sender", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Provider: ProviderAPNsFCM,
			APNs:     &APNsConfig{AuthKeyPath: filepath.Join(t.TempDir(), "nonexistent.p8"), KeyID: "x", TeamID: "x", BundleID: "x"},
			FCM:      &FCMConfig{},
		}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		// APNs init fails; FCM init typically fails in unit tests, so we get noop
		_ = sender.SendPush(ctx, "ios", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"})
	})

	T.Run("with apns_fcm provider and invalid FCM path returns iOS-only sender", func(t *testing.T) {
		t.Parallel()

		p8Path := createTestP8File(t)
		cfg := Config{
			Provider: ProviderAPNsFCM,
			APNs:     &APNsConfig{AuthKeyPath: p8Path, KeyID: "x", TeamID: "x", BundleID: "x"},
			FCM:      &FCMConfig{CredentialsPath: filepath.Join(t.TempDir(), "nonexistent.json")},
		}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		// FCM init fails, so Android not configured; should return ErrPlatformNotSupported
		err = sender.SendPush(ctx, "android", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, mobilenotifications.ErrPlatformNotSupported)
	})

	T.Run("with unknown provider returns noop", func(t *testing.T) {
		t.Parallel()

		cfg := Config{Provider: "unknown"}
		sender, err := cfg.ProvidePushSender(ctx, logger, tracer)
		require.NoError(t, err)
		require.NotNil(t, sender)
		assert.NoError(t, sender.SendPush(ctx, "ios", "token", mobilenotifications.PushMessage{Title: "title", Body: "body"}))
	})
}
