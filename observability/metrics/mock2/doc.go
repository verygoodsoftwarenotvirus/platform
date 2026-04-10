// Package mock2 provides moq-generated mock implementations of interfaces in
// the observability/metrics package. It exists alongside the hand-written
// testify-based package observability/metrics/mock and is a pilot of the
// matryer/moq workflow; consumers that want the moq style should import this
// package instead of observability/metrics/mock.
package mock2

// Regenerate via `go generate ./observability/metrics/mock2/`.

//go:generate go tool github.com/matryer/moq -out provider_mock.go -pkg mock2 -rm -fmt goimports .. Provider:ProviderMock
