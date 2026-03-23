package circuitbreaking

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/internalerrors"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/metrics"

	circuit "github.com/rubyist/circuitbreaker"
)

// ErrCircuitBroken is returned when a circuit breaker has tripped.
var ErrCircuitBroken = errors.New("service circuit broken")

// CircuitBreaker tracks failures and successes to determine whether an operation should proceed.
type CircuitBreaker interface {
	Failed()
	Succeeded()
	CanProceed() bool
	CannotProceed() bool
}

type baseImplementation struct {
	circuitBreaker *circuit.Breaker
}

func (b *baseImplementation) Failed() {
	b.circuitBreaker.Fail()
}

func (b *baseImplementation) Succeeded() {
	b.circuitBreaker.Success()
}

func (b *baseImplementation) CanProceed() bool {
	return b.circuitBreaker.Ready()
}

func (b *baseImplementation) CannotProceed() bool {
	return !b.circuitBreaker.Ready()
}

// EnsureCircuitBreaker ensures a valid CircuitBreaker is made available.
func EnsureCircuitBreaker(breaker CircuitBreaker) CircuitBreaker {
	if breaker == nil {
		slog.Info("NOOP CircuitBreaker implementation in use.")
		return NewNoopCircuitBreaker()
	}

	return breaker
}

// ProvideCircuitBreaker provides a CircuitBreaker.
func (cfg *Config) ProvideCircuitBreaker(ctx context.Context, logger logging.Logger, metricsProvider metrics.Provider) (CircuitBreaker, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("circuit breaker")
	}

	logger = logging.EnsureLogger(logger).WithValue("circuit_breaker", cfg.Name)

	if err := cfg.ValidateWithContext(ctx); err != nil {
		logger.Error("invalid config passed, providing noop circuit breaker", err)
		return NewNoopCircuitBreaker(), nil
	}

	cfg.EnsureDefaults()

	brokenCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name))
	if err != nil {
		return nil, err
	}

	failureCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name))
	if err != nil {
		return nil, err
	}

	resetCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_reset", cfg.Name))
	if err != nil {
		return nil, err
	}

	cb := circuit.NewBreakerWithOptions(&circuit.Options{
		ShouldTrip: func(cb *circuit.Breaker) bool {
			return uint64(cb.Failures()+cb.Successes()) >= cfg.MinimumSampleThreshold && cb.ErrorRate() >= cfg.ErrorRate
		},
		WindowTime:    circuit.DefaultWindowTime,
		WindowBuckets: circuit.DefaultWindowBuckets,
	})

	events := cb.Subscribe()

	go handleCircuitBreakerEvents(ctx, logger, events, failureCounter, resetCounter, brokenCounter)

	return &baseImplementation{
		circuitBreaker: cb,
	}, nil
}

func handleCircuitBreakerEvents(
	ctx context.Context,
	logger logging.Logger,
	events <-chan circuit.BreakerEvent,
	failureCounter,
	resetCounter,
	brokenCounter metrics.Int64Counter,
) {
	for be := range events {
		switch be {
		case circuit.BreakerTripped:
			brokenCounter.Add(ctx, 1)
		case circuit.BreakerReset:
			resetCounter.Add(ctx, 1)
		case circuit.BreakerFail:
			failureCounter.Add(ctx, 1)
		case circuit.BreakerReady:
			logger.Debug("circuit breaker is ready")
		}
	}
}
