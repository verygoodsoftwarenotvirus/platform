// Package mock provides moq-generated mock implementations of the interfaces
// that cache/config depends on for unit testing.
package mock

// Regenerate via `go generate ./cache/config/mock/`.

//go:generate go tool github.com/matryer/moq -out metricsprovider_mock.go -pkg mock -rm -fmt goimports ../../../observability/metrics Provider:MetricsProviderMock
