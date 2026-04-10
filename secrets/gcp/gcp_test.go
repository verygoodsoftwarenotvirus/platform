package gcp

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
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
		mc := &mockGCPClient{value: "secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)
		defer source.Close()
	})

	T.Run("with error creating lookup counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{ProjectID: "test-project"}
		source, err := NewGCPSecretSource(context.Background(), cfg, &mockGCPClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{ProjectID: "test-project"}
		source, err := NewGCPSecretSource(context.Background(), cfg, &mockGCPClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		noopMP := metrics.NewNoopMetricsProvider()
		h, histErr := noopMP.NewFloat64Histogram("test")
		require.NoError(t, histErr)

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewFloat64Histogram", name+"_latency_ms", []metric.Float64HistogramOption(nil)).Return(h, errors.New("arbitrary"))

		cfg := &Config{ProjectID: "test-project"}
		source, err := NewGCPSecretSource(context.Background(), cfg, &mockGCPClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestGCPSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mc := &mockGCPClient{value: "my-secret-value"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)
		defer source.Close()

		got, err := source.GetSecret(context.Background(), "MY_SECRET")
		require.NoError(t, err)
		assert.Equal(t, "my-secret-value", got)
	})

	T.Run("error from client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mc := &mockGCPClient{err: errors.New("gcp error")}
		source, err := NewGCPSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)
		defer source.Close()

		_, err = source.GetSecret(context.Background(), "MY_SECRET")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gcp error")
	})

	T.Run("full resource name passed through", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{ProjectID: "test-project"}
		mc := &mockGCPClient{value: "full-name-secret"}
		source, err := NewGCPSecretSource(context.Background(), cfg, mc, nil, nil, nil)
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
		mc := &mockGCPClient{}
		source, err := NewGCPSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		err = source.Close()
		require.NoError(t, err)
		assert.True(t, mc.closed)
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
