package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

const testKey = "blahblahblahblahblahblahblahblah"

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("aes provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: ProviderAES}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("salsa20 provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: ProviderSalsa20}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("empty provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{}
		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("invalid provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{Provider: "invalid"}
		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestProvideEncryptorDecryptor(T *testing.T) {
	T.Parallel()

	tracerProvider := tracing.NewNoopTracerProvider()
	logger := logging.NewNoopLogger()
	key := []byte(testKey)

	T.Run("aes provider", func(t *testing.T) {
		t.Parallel()

		encDec, err := ProvideEncryptorDecryptor(&Config{Provider: ProviderAES}, tracerProvider, logger, key)
		assert.NoError(t, err)
		assert.NotNil(t, encDec)
	})

	T.Run("salsa20 provider", func(t *testing.T) {
		t.Parallel()

		encDec, err := ProvideEncryptorDecryptor(&Config{Provider: ProviderSalsa20}, tracerProvider, logger, key)
		assert.NoError(t, err)
		assert.NotNil(t, encDec)
	})

	T.Run("empty provider defaults to salsa20", func(t *testing.T) {
		t.Parallel()

		encDec, err := ProvideEncryptorDecryptor(&Config{}, tracerProvider, logger, key)
		assert.NoError(t, err)
		assert.NotNil(t, encDec)
	})

	T.Run("invalid provider defaults to salsa20", func(t *testing.T) {
		t.Parallel()

		encDec, err := ProvideEncryptorDecryptor(&Config{Provider: "invalid"}, tracerProvider, logger, key)
		assert.NoError(t, err)
		assert.NotNil(t, encDec)
	})
}
