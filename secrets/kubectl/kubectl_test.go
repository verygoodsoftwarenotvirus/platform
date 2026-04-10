package kubectl

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewKubectlSecretSource(T *testing.T) {
	T.Parallel()

	T.Run("nil config returns error", func(t *testing.T) {
		t.Parallel()
		source, err := NewKubectlSecretSource(context.Background(), nil, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "config is required")
	})

	T.Run("missing namespace returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{}
		source, err := NewKubectlSecretSource(context.Background(), cfg, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
	})

	T.Run("with mock client succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)
	})

	T.Run("with error creating lookup counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{Namespace: "default"}
		source, err := NewKubectlSecretSource(context.Background(), cfg, &mockSecretGetter{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{Namespace: "default"}
		source, err := NewKubectlSecretSource(context.Background(), cfg, &mockSecretGetter{}, nil, nil, mp)
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

		cfg := &Config{Namespace: "default"}
		source, err := NewKubectlSecretSource(context.Background(), cfg, &mockSecretGetter{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestKubectlSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"password": []byte("s3cret"),
				},
			},
		}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "db-creds/password")
		require.NoError(t, err)
		assert.Equal(t, "s3cret", got)
		assert.Equal(t, "db-creds", mc.lastName)
	})

	T.Run("missing slash in name", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		_, err = source.GetSecret(context.Background(), "no-slash")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected format")
	})

	T.Run("key not found", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{
			secret: &corev1.Secret{
				Data: map[string][]byte{
					"username": []byte("admin"),
				},
			},
		}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		_, err = source.GetSecret(context.Background(), "db-creds/password")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key \"password\" not found")
	})

	T.Run("client error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{err: errors.New("k8s api error")}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		_, err = source.GetSecret(context.Background(), "db-creds/password")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "k8s api error")
	})
}

func TestKubectlSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Namespace: "default"}
		mc := &mockSecretGetter{}
		source, err := NewKubectlSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		err = source.Close()
		require.NoError(t, err)
	})
}

func TestResolveName(T *testing.T) {
	T.Parallel()

	T.Run("valid name", func(t *testing.T) {
		t.Parallel()
		secretName, key, err := resolveName("my-secret/my-key")
		require.NoError(t, err)
		assert.Equal(t, "my-secret", secretName)
		assert.Equal(t, "my-key", key)
	})

	T.Run("name with multiple slashes", func(t *testing.T) {
		t.Parallel()
		secretName, key, err := resolveName("my-secret/nested/key")
		require.NoError(t, err)
		assert.Equal(t, "my-secret", secretName)
		assert.Equal(t, "nested/key", key)
	})

	T.Run("no slash", func(t *testing.T) {
		t.Parallel()
		_, _, err := resolveName("no-slash")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected format")
	})
}

type mockSecretGetter struct {
	secret   *corev1.Secret
	err      error
	lastName string
}

func (m *mockSecretGetter) Get(_ context.Context, name string, _ metav1.GetOptions) (*corev1.Secret, error) {
	m.lastName = name
	if m.err != nil {
		return nil, m.err
	}
	return m.secret, nil
}
