package secretscfg

import (
	"context"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/secrets"
	"github.com/verygoodsoftwarenotvirus/platform/secrets/env"
	"github.com/verygoodsoftwarenotvirus/platform/secrets/gcp"
	"github.com/verygoodsoftwarenotvirus/platform/secrets/ssm"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProviderEnv represents environment variables.
	ProviderEnv = "env"
	// ProviderNoop represents the noop provider.
	ProviderNoop = "noop"
	// ProviderGCP represents GCP Secret Manager.
	ProviderGCP = "gcp"
	// ProviderSSM represents AWS SSM Parameter Store.
	ProviderSSM = "ssm"
)

// Config configures secret source selection.
type Config struct {
	GCPClient gcp.SecretVersionAccessor `json:"-"`
	SSMClient ssm.GetParameterAPI       `json:"-"`
	Env       *env.Config               `env:"init"     envPrefix:"ENV_" json:"env,omitempty"`
	GCP       *gcp.Config               `env:"init"     envPrefix:"GCP_" json:"gcp,omitempty"`
	SSM       *ssm.Config               `env:"init"     envPrefix:"SSM_" json:"ssm,omitempty"`
	Provider  string                    `env:"PROVIDER" json:"provider"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the config.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Provider, validation.In(ProviderEnv, ProviderNoop, ProviderGCP, ProviderSSM, "")),
		validation.Field(&cfg.GCP, validation.When(cfg.Provider == ProviderGCP, validation.Required), validation.When(cfg.Provider != ProviderGCP, validation.Nil)),
		validation.Field(&cfg.SSM, validation.When(cfg.Provider == ProviderSSM, validation.Required), validation.When(cfg.Provider != ProviderSSM, validation.Nil)),
	)
}

// ProvideSecretSource returns a SecretSource from config.
func (cfg *Config) ProvideSecretSource(ctx context.Context) (secrets.SecretSource, error) {
	if cfg == nil {
		return env.NewEnvSecretSource(), nil
	}

	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))
	switch provider {
	case "", ProviderEnv:
		return env.NewEnvSecretSource(), nil
	case ProviderNoop:
		return secrets.NewNoopSecretSource(), nil
	case ProviderGCP:
		if cfg.GCP == nil {
			return nil, fmt.Errorf("gcp provider requires gcp config")
		}
		return gcp.NewGCPSecretSource(ctx, cfg.GCP, cfg.GCPClient)
	case ProviderSSM:
		if cfg.SSM == nil {
			return nil, fmt.Errorf("ssm provider requires ssm config")
		}
		return ssm.NewSSMSecretSource(ctx, cfg.SSM, cfg.SSMClient)
	default:
		return nil, fmt.Errorf("unknown secret source provider: %q", cfg.Provider)
	}
}
