package mockmetrics

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/metric"
)

type MetricsProvider struct {
	mock.Mock
}

// NewFloat64Counter is a mock method.
func (m *MetricsProvider) NewFloat64Counter(name string, options ...metric.Float64CounterOption) (metrics.Float64Counter, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Float64Counter), args.Error(1)
}

// NewFloat64Gauge is a mock method.
func (m *MetricsProvider) NewFloat64Gauge(name string, options ...metric.Float64GaugeOption) (metrics.Float64Gauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Float64Gauge), args.Error(1)
}

// NewFloat64UpDownCounter is a mock method.
func (m *MetricsProvider) NewFloat64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metrics.Float64UpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Float64UpDownCounter), args.Error(1)
}

// NewFloat64Histogram is a mock method.
func (m *MetricsProvider) NewFloat64Histogram(name string, options ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Float64Histogram), args.Error(1)
}

// NewInt64Counter is a mock method.
func (m *MetricsProvider) NewInt64Counter(name string, options ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Int64Counter), args.Error(1)
}

// NewInt64Gauge is a mock method.
func (m *MetricsProvider) NewInt64Gauge(name string, options ...metric.Int64GaugeOption) (metrics.Int64Gauge, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Int64Gauge), args.Error(1)
}

// NewInt64UpDownCounter is a mock method.
func (m *MetricsProvider) NewInt64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metrics.Int64UpDownCounter, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Int64UpDownCounter), args.Error(1)
}

// NewInt64Histogram is a mock method.
func (m *MetricsProvider) NewInt64Histogram(name string, options ...metric.Int64HistogramOption) (metrics.Int64Histogram, error) {
	args := m.Called(name, options)
	return args.Get(0).(metrics.Int64Histogram), args.Error(1)
}

// Shutdown is a mock method.
func (m *MetricsProvider) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MeterProvider is a mock method.
func (m *MetricsProvider) MeterProvider() metric.MeterProvider {
	args := m.Called()
	return args.Get(0).(metric.MeterProvider)
}
