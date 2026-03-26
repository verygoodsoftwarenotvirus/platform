package env

import (
	"context"
	"os"

	"github.com/verygoodsoftwarenotvirus/platform/v3/secrets"
)

type envSecretSource struct{}

// NewEnvSecretSource returns a SecretSource that reads from environment variables.
func NewEnvSecretSource() secrets.SecretSource {
	return &envSecretSource{}
}

func (e *envSecretSource) GetSecret(ctx context.Context, name string) (string, error) {
	return os.Getenv(name), nil
}

func (e *envSecretSource) Close() error {
	return nil
}
