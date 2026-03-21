package gcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func TestNewGCPSecretSource(t *testing.T) {
	t.Parallel()

	t.Run("nil config returns error", func(t *testing.T) {
		t.Parallel()
		source, err := NewGCPSecretSource(context.Background(), nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "config is required")
	})

	t.Run("missing ProjectID returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: ""}
		source, err := NewGCPSecretSource(context.Background(), cfg, nil)
		require.Error(t, err)
		assert.Nil(t, source)
	})

	t.Run("with mock client succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)
		require.NotNil(t, source)
		defer source.Close()
	})
}

func TestGCPSecretSource_GetSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "my-secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)
		defer source.Close()

		got, err := source.GetSecret(context.Background(), "MY_SECRET")
		require.NoError(t, err)
		assert.Equal(t, "my-secret-value", got)
	})

	t.Run("error from client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{err: errors.New("gcp error")}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)
		defer source.Close()

		_, err = source.GetSecret(context.Background(), "MY_SECRET")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gcp error")
	})

	t.Run("full resource name passed through", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mock := &mockGCPClient{value: "full-name-secret"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mock)
		require.NoError(t, err)
		defer source.Close()

		got, err := source.GetSecret(context.Background(), "projects/other-project/secrets/foo/versions/latest")
		require.NoError(t, err)
		assert.Equal(t, "full-name-secret", got)
	})
}

func TestGCPSecretSource_Close(t *testing.T) {
	t.Parallel()

	cfg := &Config{ProjectID: "test-project"}
	mock := &mockGCPClient{}
	source, err := NewGCPSecretSource(context.Background(), cfg, mock)
	require.NoError(t, err)

	err = source.Close()
	require.NoError(t, err)
	assert.True(t, mock.closed)
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
