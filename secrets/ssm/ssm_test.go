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
	"github.com/shoenig/test"
	"github.com/stretchr/testify/assert"
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

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				test.EqOp(t, name+"_lookups", counterName)
				return metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary")
			},
		}

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating error counter", func(t *testing.T) {
		t.Parallel()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch counterName {
				case name + "_lookups":
					return metrics.Int64CounterForTest(t, "x"), nil
				case name + "_errors":
					return metrics.Int64CounterForTest(t, "x"), errors.New("arbitrary")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", counterName)
				return nil, nil
			},
		}

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
	})

	T.Run("with error creating latency histogram", func(t *testing.T) {
		t.Parallel()

		noopMP := metrics.NewNoopMetricsProvider()
		h, histErr := noopMP.NewFloat64Histogram("test")
		require.NoError(t, histErr)

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				return metrics.Int64CounterForTest(t, "x"), nil
			},
			NewFloat64HistogramFunc: func(histName string, _ ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
				test.EqOp(t, name+"_latency_ms", histName)
				return h, errors.New("arbitrary")
			},
		}

		cfg := &Config{Region: "us-east-1"}
		source, err := NewSSMSecretSource(context.Background(), cfg, &mockSSMClient{}, nil, nil, mp)
		require.Error(t, err)
		assert.Nil(t, source)

		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
		test.SliceLen(t, 1, mp.NewFloat64HistogramCalls())
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
