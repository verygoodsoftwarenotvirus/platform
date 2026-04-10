package ssm

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

func TestNewSSMSecretSource(T *testing.T) {
	T.Parallel()

	T.Run("nil config returns error", func(t *testing.T) {
		t.Parallel()
		source, err := NewSSMSecretSource(context.Background(), nil, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "config is required")
	})

	T.Run("missing Region returns error", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: ""}
		source, err := NewSSMSecretSource(context.Background(), cfg, nil, nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, source)
	})

	T.Run("with mock client succeeds", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mc := &mockSSMClient{value: "param-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, source)
	})

	T.Run("with error creating lookup counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.MetricsProvider{}
		mp.On("NewInt64Counter", name+"_lookups", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), nil)
		mp.On("NewInt64Counter", name+"_errors", []metric.Int64CounterOption(nil)).Return(metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary"))

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
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

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

func TestSSMSecretSource_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("success", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mc := &mockSSMClient{value: "my-param-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "MY_PARAM")
		require.NoError(t, err)
		assert.Equal(t, "my-param-value", got)
	})

	T.Run("error from client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1"}
		mc := &mockSSMClient{err: errors.New("ssm error")}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		_, err = source.GetSecret(context.Background(), "MY_PARAM")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssm error")
	})

	T.Run("name with prefix", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1", Prefix: "/myapp/"}
		mc := &mockSSMClient{value: "prefixed-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "MY_PARAM")
		require.NoError(t, err)
		assert.Equal(t, "prefixed-value", got)
		assert.Equal(t, "/myapp/MY_PARAM", mc.lastName)
	})

	T.Run("name already path", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{Region: "us-east-1", Prefix: "/myapp/"}
		mc := &mockSSMClient{value: "path-value"}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		got, err := source.GetSecret(context.Background(), "/existing/path/param")
		require.NoError(t, err)
		assert.Equal(t, "path-value", got)
		assert.Equal(t, "/existing/path/param", mc.lastName)
	})
}

func TestSSMSecretSource_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Region: "us-east-1"}
		mc := &mockSSMClient{}
		source, err := NewSSMSecretSource(context.Background(), cfg, mc, nil, nil, nil)
		require.NoError(t, err)

		err = source.Close()
		require.NoError(t, err)
	})
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
