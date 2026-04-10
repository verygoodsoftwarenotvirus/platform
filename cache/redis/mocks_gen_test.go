package redis

// Regenerate mocks via `go generate ./cache/redis/...`.

//go:generate go tool github.com/matryer/moq -out redisclient_mock_test.go -pkg redis -rm -fmt goimports . redisClient
//go:generate go tool github.com/matryer/moq -out metricsprovider_mock_test.go -pkg redis -rm -fmt goimports ../../observability/metrics Provider
//go:generate go tool github.com/matryer/moq -out circuitbreaker_mock_test.go -pkg redis -rm -fmt goimports ../../circuitbreaking CircuitBreaker
