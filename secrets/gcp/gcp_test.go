package gcp

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGCPSecretSource(T *testing.T) {
	T.Parallel()

	T.Run("nil config returns error", func(t *testing.T) {
		t.Parallel()
		source, err := NewGCPSecretSource(context.Background(), nil, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "config is required")
	})

	T.Run("missing ProjectID returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: ""}
		source, err := NewGCPSecretSource(context.Background(), cfg, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
	})

	T.Run("with mock client succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)
		defer source.Close()
	})
}

func TestGCPSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "my-secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock, nil, nil, nil)
		require.NoError(t, err)
		defer source.Close()

		got, err := source.GetSecret(context.Background(), "MY_SECRET")
		require.NoError(t, err)
		assert.Equal(t, "my-secret-value", got)
	})

	T.Run("error from client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{err: errors.New("gcp error")}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock, nil, nil, nil)
		require.NoError(t, err)
		defer source.Close()

		_, err = source.GetSecret(context.Background(), "MY_SECRET")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gcp error")
	})

	T.Run("full resource name passed through", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "full-name-secret"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock, nil, nil, nil)
		require.NoError(t, err)
		defer source.Close()

		got, err := source.GetSecret(context.Background(), "projects/other-project/secrets/foo/versions/latest")
		require.NoError(t, err)
		assert.Equal(t, "full-name-secret", got)
	})
}

func TestGCPSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock, nil, nil, nil)
		require.NoError(t, err)

		err = source.Close()
		require.NoError(t, err)
		assert.True(t, mock.closed)
	})
}

type mockGCPClient struct {
	err    error
	value  string
	closed bool
}

func (m *mockGCPClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(m.value)},
	}, nil
}

func (m *mockGCPClient) Close() error {
	m.closed = true
	return nil
}
