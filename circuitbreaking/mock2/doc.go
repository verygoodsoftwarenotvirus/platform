// Package mock2 provides moq-generated mock implementations of interfaces in
// the circuitbreaking package. It exists alongside the hand-written
// testify-based package circuitbreaking/mock and is a pilot of the
// matryer/moq workflow; consumers that want the moq style should import this
// package instead of circuitbreaking/mock.
package mock2

// Regenerate via `go generate ./circuitbreaking/mock2/`.

//go:generate go tool github.com/matryer/moq -out circuitbreaker_mock.go -pkg mock2 -rm -fmt goimports .. CircuitBreaker:CircuitBreakerMock
