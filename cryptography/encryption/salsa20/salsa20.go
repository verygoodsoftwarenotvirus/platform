package salsa20

import (
	"github.com/verygoodsoftwarenotvirus/platform/v5/cryptography/encryption"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

// salsa20Impl is the standard EncryptorDecryptor implementation.
type salsa20Impl struct {
	tracer tracing.Tracer
	logger logging.Logger
	key    [32]byte
}

func NewEncryptorDecryptor(tracerProvider tracing.TracerProvider, logger logging.Logger, key []byte) (encryption.EncryptorDecryptor, error) {
	if len(key) != 32 {
		return nil, encryption.ErrIncorrectKeyLength
	}

	var key32 [32]byte
	copy(key32[:], key)

	return &salsa20Impl{
		logger: logging.NewNamedLogger(logger, "encryptor"),
		tracer: tracing.NewNamedTracer(tracerProvider, "encryptor"),
		key:    key32,
	}, nil
}
