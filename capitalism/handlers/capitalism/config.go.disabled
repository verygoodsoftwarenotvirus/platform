package capitalism

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures the service.
type Config struct {
	_ struct{} `json:"-"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
	)
}
