package circuitbreaking

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v2/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/metric"
)

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestProvideCircuitBreaker(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()

		cb, err := ProvideCircuitBreaker(ctx, cfg, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)
	})

	T.Run("with error providing first metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()
		i64Counter := &mockmetrics.Int64Counter{}

		mp := &mockmetrics.MetricsProvider{}
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, errors.New("arbitrary"))

		cb, err := ProvideCircuitBreaker(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error providing second metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()
		i64Counter := &mockmetrics.Int64Counter{}

		mp := &mockmetrics.MetricsProvider{}
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, nil)
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, errors.New("arbitrary"))

		cb, err := ProvideCircuitBreaker(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mp)
	})

	T.Run("with error providing third metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()
		i64Counter := &mockmetrics.Int64Counter{}

		mp := &mockmetrics.MetricsProvider{}
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, nil)
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, nil)
		mp.On(reflection.GetMethodName(mp.NewInt64Counter), fmt.Sprintf("%s_circuit_breaker_reset", cfg.Name), []metric.Int64CounterOption(nil)).Return(i64Counter, errors.New("arbitrary"))

		cb, err := ProvideCircuitBreaker(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mp)
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestEnsureCircuitBreaker(T *testing.T) {
	T.Run("with nil breaker", func(t *testing.T) {
		actual := EnsureCircuitBreaker(nil)
		assert.NotNil(t, actual)
		assert.IsType(t, &NoopCircuitBreaker{}, actual)
	})

	T.Run("with non-nil breaker", func(t *testing.T) {
		input := NewNoopCircuitBreaker()
		actual := EnsureCircuitBreaker(input)
		assert.Equal(t, input, actual)
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestConfig_ProvideCircuitBreaker(T *testing.T) {
	T.Run("with nil config", func(t *testing.T) {
		ctx := t.Context()

		var cfg *Config
		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.Nil(t, cb)
		assert.Error(t, err)
	})

	T.Run("with invalid config", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{
			Name:      "",
			ErrorRate: 200,
		}

		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)
		assert.IsType(t, &NoopCircuitBreaker{}, cb)
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestBaseImplementation(T *testing.T) {
	T.Run("Failed", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              99,
			MinimumSampleThreshold: 1000,
		}

		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)

		cb.Failed()
	})

	T.Run("Succeeded", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              99,
			MinimumSampleThreshold: 1000,
		}

		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)

		cb.Succeeded()
	})

	T.Run("CanProceed", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              99,
			MinimumSampleThreshold: 1000,
		}

		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)

		assert.True(t, cb.CanProceed())
	})

	T.Run("CannotProceed", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              99,
			MinimumSampleThreshold: 1000,
		}

		cb, err := cfg.ProvideCircuitBreaker(ctx, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)

		assert.False(t, cb.CannotProceed())
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestCircuitBreaker_Integration(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		t.SkipNow() // cannot run this with the race detector on

		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              1,
			MinimumSampleThreshold: 1,
		}

		cb, err := ProvideCircuitBreaker(ctx, cfg, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)

		assert.True(t, cb.CanProceed())
		cb.Failed()
		assert.True(t, cb.CannotProceed())
		cb.Succeeded()
		assert.Eventually(
			t,
			func() bool {
				return cb.CanProceed()
			},
			5*time.Second,
			500*time.Millisecond,
		)
	})
}
