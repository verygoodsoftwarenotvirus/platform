package ssm

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/secrets"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// GetParameterAPI abstracts GetParameter for testability.
type GetParameterAPI interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type ssmSecretSource struct {
	client GetParameterAPI
	prefix string
}

// NewSSMSecretSource creates a SecretSource backed by AWS SSM Parameter Store.
// If client is nil, a new client is created using the default credential chain.
func NewSSMSecretSource(ctx context.Context, cfg *Config, client GetParameterAPI) (secrets.SecretSource, error) {
	if cfg == nil {
		return nil, errors.New("ssm secret source: config is required")
	}
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, errors.Wrap(err, "ssm secret source")
	}

	if client != nil {
		return &ssmSecretSource{
			client: client,
			prefix: cfg.Prefix,
		}, nil
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, errors.Wrap(err, "ssm secret source: loading aws config")
	}

	return &ssmSecretSource{
		client: ssm.NewFromConfig(awsCfg),
		prefix: cfg.Prefix,
	}, nil
}

func (s *ssmSecretSource) GetSecret(ctx context.Context, name string) (string, error) {
	paramName := s.resolveName(name)
	input := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	}

	output, err := s.client.GetParameter(ctx, input)
	if err != nil {
		return "", errors.Wrapf(err, "getting parameter %q", name)
	}
	if output.Parameter == nil {
		return "", nil
	}
	return aws.ToString(output.Parameter.Value), nil
}

func (s *ssmSecretSource) Close() error {
	return nil
}

func (s *ssmSecretSource) resolveName(name string) string {
	if strings.HasPrefix(name, "/") {
		return name
	}
	if s.prefix != "" {
		return s.prefix + name
	}
	return name
}

// Ensure ssmSecretSource implements secrets.SecretSource.
var _ secrets.SecretSource = (*ssmSecretSource)(nil)
