package secretscfg

import (
	"context"
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/secrets/gcp"
	"github.com/verygoodsoftwarenotvirus/platform/v5/secrets/kubectl"
	"github.com/verygoodsoftwarenotvirus/platform/v5/secrets/ssm"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockGCPClient struct {
	value string
}

func (m *mockGCPClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(m.value)},
	}, nil
}

func (m *mockGCPClient) Close() error { return nil }

type mockSSMClient struct {
	value string
}

func (m *mockSSMClient) GetParameter(ctx context.Context, params *awsssm.GetParameterInput, optFns ...func(*awsssm.Options)) (*awsssm.GetParameterOutput, error) {
	return &awsssm.GetParameterOutput{
		Parameter: &types.Parameter{
			Value: aws.String(m.value),
		},
	}, nil
}

type mockKubectlClient struct {
	secret *corev1.Secret
}

func (m *mockKubectlClient) Get(_ context.Context, _ string, _ metav1.GetOptions) (*corev1.Secret, error) {
	return m.secret, nil
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid env provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderEnv}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("valid noop provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderNoop}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("valid gcp provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderGCP, GCP: &gcp.Config{ProjectID: "my-project"}}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("invalid gcp provider missing config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderGCP}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("valid ssm provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderSSM, SSM: &ssm.Config{Region: "us-east-1"}}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("invalid ssm provider missing config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderSSM}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("valid kubectl provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderKubectl, Kubectl: &kubectl.Config{Namespace: "default"}}
		require.NoError(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("invalid kubectl provider missing config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: ProviderKubectl}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})

	T.Run("unknown provider", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Provider: "vault"}
		require.Error(t, cfg.ValidateWithContext(context.Background()))
	})
}

func TestConfig_ProvideSecretSource(T *testing.T) {
	T.Parallel()

	T.Run("nil config returns env source", func(t *testing.T) {
		t.Parallel()

		var cfg *Config
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_NIL_CONFIG_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	T.Run("empty provider returns env source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ""}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_EMPTY_PROVIDER_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	T.Run("env provider returns env source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderEnv}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		key := "TEST_ENV_PROVIDER_" + t.Name()
		value := "from-env"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		got, err := source.GetSecret(context.Background(), key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	T.Run("noop provider returns noop source", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderNoop}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		got, err := source.GetSecret(context.Background(), "any")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	T.Run("gcp provider with mock client", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:  ProviderGCP,
			GCP:       &gcp.Config{ProjectID: "test-project"},
			GCPClient: &mockGCPClient{value: "gcp-secret-value"},
		}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		got, err := source.GetSecret(context.Background(), "MY_SECRET")
		require.NoError(t, err)
		assert.Equal(t, "gcp-secret-value", got)
	})

	T.Run("ssm provider with mock client", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider:  ProviderSSM,
			SSM:       &ssm.Config{Region: "us-east-1"},
			SSMClient: &mockSSMClient{value: "ssm-param-value"},
		}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		got, err := source.GetSecret(context.Background(), "MY_PARAM")
		require.NoError(t, err)
		assert.Equal(t, "ssm-param-value", got)
	})

	T.Run("kubectl provider with mock client", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Provider: ProviderKubectl,
			Kubectl:  &kubectl.Config{Namespace: "default"},
			KubectlClient: &mockKubectlClient{
				secret: &corev1.Secret{
					Data: map[string][]byte{
						"password": []byte("k8s-secret-value"),
					},
				},
			},
		}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)

		got, err := source.GetSecret(context.Background(), "my-secret/password")
		require.NoError(t, err)
		assert.Equal(t, "k8s-secret-value", got)
	})

	T.Run("unknown provider returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: "vault"}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "unknown")
	})

	T.Run("gcp provider with nil gcp config returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderGCP}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "gcp")
	})

	T.Run("ssm provider with nil ssm config returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderSSM}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "ssm")
	})

	T.Run("kubectl provider with nil kubectl config returns error", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Provider: ProviderKubectl}
		source, err := cfg.ProvideSecretSource(context.Background(), nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "kubectl")
	})
}
