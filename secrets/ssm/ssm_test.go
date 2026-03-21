package ssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSMSecretSource(t *testing.T) {
	t.Parallel()

	t.Run("nil config returns error", func(t *testing.T) {
		t.Parallel()
		source, err := NewSSMSecretSource(context.Background(), nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "config is required")
	})

	t.Run("missing Region returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: ""}
		source, err := NewSSMSecretSource(context.Background(), cfg, nil)
		require.Error(t, err)
		assert.Nil(t, source)
	})

	t.Run("with mock client succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mock := &mockSSMClient{value: "param-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)
		require.NotNil(t, source)
	})
}

func TestSSMSecretSource_GetSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mock := &mockSSMClient{value: "my-param-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "MY_PARAM")
		require.NoError(t, err)
		assert.Equal(t, "my-param-value", got)
	})

	t.Run("error from client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mock := &mockSSMClient{err: errors.New("ssm error")}
		source, err := NewSSMSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)

		_, err = source.GetSecret(context.Background(), "MY_PARAM")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssm error")
	})

	t.Run("name with prefix", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1", Prefix: "/myapp/"}
		mock := &mockSSMClient{value: "prefixed-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "MY_PARAM")
		require.NoError(t, err)
		assert.Equal(t, "prefixed-value", got)
		assert.Equal(t, "/myapp/MY_PARAM", mock.lastName)
	})

	t.Run("name already path", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1", Prefix: "/myapp/"}
		mock := &mockSSMClient{value: "path-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "/existing/path/param")
		require.NoError(t, err)
		assert.Equal(t, "path-value", got)
		assert.Equal(t, "/existing/path/param", mock.lastName)
	})
}

func TestSSMSecretSource_Close(t *testing.T) {
	t.Parallel()

	cfg := &Config{Region: "us-east-1"}
	mock := &mockSSMClient{}
	source, err := NewSSMSecretSource(context.Background(), cfg, mock)
	require.NoError(t, err)

	err = source.Close()
	require.NoError(t, err)
}

type mockSSMClient struct {
	value    string
	err      error
	lastName string
}

func (m *mockSSMClient) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	if params.Name != nil {
		m.lastName = aws.ToString(params.Name)
	}
	if m.err != nil {
		return nil, m.err
	}
	return &ssm.GetParameterOutput{
		Parameter: &types.Parameter{
			Value: aws.String(m.value),
		},
	}, nil
}
