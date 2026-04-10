package env

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v5/secrets"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

var _ secrets.SecretSource = (*envSecretSource)(nil)

func TestNewEnvSecretSource(T *testing.T) {
	T.Parallel()

	T.Run("with error creating lookup counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		source, err := NewEnvSecretSource(nil, nil, mp)
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
		mp.On("NewFloat64Histogram", name+"_latency_ms", []metric.Float64HistogramOption(nil)).Return(h, errors.New("arbitrary"))

		source, err := NewEnvSecretSource(nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestEnvSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("returns set env var", func(t *testing.T) {
		t.Parallel()

		key := "TEST_SECRET_" + t.Name()
		value := "secret-value"
		require.NoError(t, os.Setenv(key, value))
		t.Cleanup(func() { _ = os.Unsetenv(key) })

		source, err := NewEnvSecretSource(nil, nil, nil)
		require.NoError(t, err)
		ctx := context.Background()

		got, err := source.GetSecret(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, got)
	})

	T.Run("returns empty for unset env var", func(t *testing.T) {
		t.Parallel()

		key := "TEST_SECRET_UNSET_" + t.Name()
		require.NoError(t, os.Unsetenv(key))

		source, err := NewEnvSecretSource(nil, nil, nil)
		require.NoError(t, err)
		ctx := context.Background()

		got, err := source.GetSecret(ctx, key)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestEnvSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source, err := NewEnvSecretSource(nil, nil, nil)
		require.NoError(t, err)

		err = source.Close()
		require.NoError(t, err)
	})
}
