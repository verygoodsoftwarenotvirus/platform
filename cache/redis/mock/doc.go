// Package mock provides moq-generated mock implementations of the interfaces
// that cache/redis depends on for unit testing: the internal redisClient test
// seam plus metrics.Provider and circuitbreaking.CircuitBreaker.
package mock

// Regenerate via `go generate ./cache/redis/mock/`.

//go:generate go tool github.com/matryer/moq -out redisclient_mock.go -pkg mock -rm -skip-ensure -fmt goimports .. redisClient:RedisClientMock
//go:generate go tool github.com/matryer/moq -out metricsprovider_mock.go -pkg mock -rm -fmt goimports ../../../observability/metrics Provider:MetricsProviderMock
//go:generate go tool github.com/matryer/moq -out circuitbreaker_mock.go -pkg mock -rm -fmt goimports ../../../circuitbreaking CircuitBreaker:CircuitBreakerMock
