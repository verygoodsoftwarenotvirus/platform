package metrics

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/metric"
)

var _ Provider = (*MockProvider)(nil)

type MockProvider struct {
	mock.Mock
}

// NewFloat64Counter satisfies our interface.
func (m *MockProvider) NewFloat64Counter(name string, options ...metric.Float64CounterOption) (Float64Counter, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Float64Counter), returnValues.Error(1)
}

// NewFloat64Gauge satisfies our interface.
func (m *MockProvider) NewFloat64Gauge(name string, options ...metric.Float64GaugeOption) (Float64Gauge, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Float64Gauge), returnValues.Error(1)
}

// NewFloat64UpDownCounter satisfies our interface.
func (m *MockProvider) NewFloat64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (Float64UpDownCounter, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Float64UpDownCounter), returnValues.Error(1)
}

// NewFloat64Histogram satisfies our interface.
func (m *MockProvider) NewFloat64Histogram(name string, options ...metric.Float64HistogramOption) (Float64Histogram, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Float64Histogram), returnValues.Error(1)
}

// NewInt64Counter satisfies our interface.
func (m *MockProvider) NewInt64Counter(name string, options ...metric.Int64CounterOption) (Int64Counter, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Int64Counter), returnValues.Error(1)
}

// NewInt64Gauge satisfies our interface.
func (m *MockProvider) NewInt64Gauge(name string, options ...metric.Int64GaugeOption) (Int64Gauge, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Int64Gauge), returnValues.Error(1)
}

// NewInt64UpDownCounter satisfies our interface.
func (m *MockProvider) NewInt64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (Int64UpDownCounter, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Int64UpDownCounter), returnValues.Error(1)
}

// NewInt64Histogram satisfies our interface.
func (m *MockProvider) NewInt64Histogram(name string, options ...metric.Int64HistogramOption) (Int64Histogram, error) {
	returnValues := m.Called(name, options)
	return returnValues.Get(0).(metric.Int64Histogram), returnValues.Error(1)
}

// Shutdown satisfies our interface.
func (m *MockProvider) Shutdown(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// MeterProvider satisfies our interface.
func (m *MockProvider) MeterProvider() metric.MeterProvider {
	return m.Called().Get(0).(metric.MeterProvider)
}
