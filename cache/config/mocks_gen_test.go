package config

// Regenerate mocks via `go generate ./cache/config/...`.

//go:generate go tool github.com/matryer/moq -out metricsprovider_mock_test.go -pkg config -rm -fmt goimports ../../observability/metrics Provider
