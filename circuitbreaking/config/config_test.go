package circuitbreakingcfg

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics"
	mockmetrics "github.com/verygoodsoftwarenotvirus/platform/v5/observability/metrics/mock"

	circuit "github.com/rubyist/circuitbreaker"
	"github.com/shoenig/test"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:                   t.Name(),
			ErrorRate:              0.99,
			MinimumSampleThreshold: 123,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})

	T.Run("with missing name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:      "",
			ErrorRate: 0.99,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.Error(t, err)
	})

	T.Run("with error rate exceeding max", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cfg := &Config{
			Name:      t.Name(),
			ErrorRate: 200,
		}

		err := cfg.ValidateWithContext(ctx)
		assert.Error(t, err)
	})
}

func TestConfig_EnsureDefaults(T *testing.T) {
	T.Parallel()

	T.Run("with empty config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{}
		cfg.EnsureDefaults()

		assert.Equal(t, "UNKNOWN", cfg.Name)
		assert.Equal(t, float64(100), cfg.ErrorRate)
		assert.Equal(t, uint64(1_000_000), cfg.MinimumSampleThreshold)
	})

	T.Run("does not override set values", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Name:                   "test",
			ErrorRate:              50.0,
			MinimumSampleThreshold: 500,
		}
		cfg.EnsureDefaults()

		assert.Equal(t, "test", cfg.Name)
		assert.Equal(t, 50.0, cfg.ErrorRate)
		assert.Equal(t, uint64(500), cfg.MinimumSampleThreshold)
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestProvideCircuitBreakerFromConfig(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()

		cb, err := ProvideCircuitBreakerFromConfig(ctx, cfg, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
		assert.NotNil(t, cb)
		assert.NoError(t, err)
	})

	T.Run("with error providing first metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				test.EqOp(t, fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name), counterName)
				return &mockmetrics.Int64CounterMock{}, errors.New("arbitrary")
			},
		}

		cb, err := ProvideCircuitBreakerFromConfig(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		test.SliceLen(t, 1, mp.NewInt64CounterCalls())
	})

	T.Run("with error providing second metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch counterName {
				case fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name):
					return &mockmetrics.Int64CounterMock{}, nil
				case fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name):
					return &mockmetrics.Int64CounterMock{}, errors.New("arbitrary")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", counterName)
				return nil, nil
			},
		}

		cb, err := ProvideCircuitBreakerFromConfig(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		test.SliceLen(t, 2, mp.NewInt64CounterCalls())
	})

	T.Run("with error providing third metric", func(t *testing.T) {
		cfg := &Config{}
		cfg.EnsureDefaults()

		ctx := t.Context()

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch counterName {
				case fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name),
					fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name):
					return &mockmetrics.Int64CounterMock{}, nil
				case fmt.Sprintf("%s_circuit_breaker_reset", cfg.Name):
					return &mockmetrics.Int64CounterMock{}, errors.New("arbitrary")
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", counterName)
				return nil, nil
			},
		}

		cb, err := ProvideCircuitBreakerFromConfig(ctx, cfg, logging.NewNoopLogger(), mp)
		assert.Nil(t, cb)
		assert.Error(t, err)

		test.SliceLen(t, 3, mp.NewInt64CounterCalls())
	})
}

//nolint:paralleltest // race condition in the core circuit breaker library, I think?
func TestEnsureCircuitBreaker(T *testing.T) {
	T.Run("with nil breaker", func(t *testing.T) {
		actual := EnsureCircuitBreaker(nil)
		assert.NotNil(t, actual)
	})

	T.Run("with non-nil breaker", func(t *testing.T) {
		input := noop.NewCircuitBreaker()
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
func TestHandleCircuitBreakerEvents(T *testing.T) {
	T.Run("handles all event types and exits on channel close", func(t *testing.T) {
		ctx := t.Context()

		i64Counter := &mockmetrics.Int64CounterMock{
			AddFunc: func(_ context.Context, _ int64, _ ...metric.AddOption) {},
		}

		mp := &mockmetrics.ProviderMock{
			NewInt64CounterFunc: func(counterName string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
				switch counterName {
				case "failure", "reset", "broken":
					return i64Counter, nil
				}
				t.Fatalf("unexpected NewInt64Counter call: %q", counterName)
				return nil, nil
			},
		}

		failure, err := mp.NewInt64Counter("failure")
		assert.NoError(t, err)
		reset, err := mp.NewInt64Counter("reset")
		assert.NoError(t, err)
		broken, err := mp.NewInt64Counter("broken")
		assert.NoError(t, err)

		events := make(chan circuit.BreakerEvent, 4)
		events <- circuit.BreakerTripped
		events <- circuit.BreakerReset
		events <- circuit.BreakerFail
		events <- circuit.BreakerReady
		close(events)

		handleCircuitBreakerEvents(ctx, logging.NewNoopLogger(), events, failure, reset, broken)

		test.SliceLen(t, 3, mp.NewInt64CounterCalls())
		test.SliceLen(t, 3, i64Counter.AddCalls())
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

		cb, err := ProvideCircuitBreakerFromConfig(ctx, cfg, logging.NewNoopLogger(), metrics.NewNoopMetricsProvider())
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
