# Code Review — 2026-03-22

Comprehensive repository audit identifying inconsistencies, missing patterns, bugs, and quality issues.

## Critical / High

### 1. `EnsureDefaults()` missing in 13 of 16 config packages

CLAUDE.md documents this as a standard pattern, but only `circuitbreaking`, `ratelimiting`, and `retry` implement it.

**Missing in:** analytics, cache, capitalism, database, email, eventstream, featureflags, llm, messagequeue, notifications, routing, secrets, uploads.

### 2. Error handling doesn't use the platform `errors` package

The repo wraps `cockroachdb/errors` in its own `errors/` package with sentinels, but actual usage is limited to ~3 files. The vast majority of the codebase (~223 instances) uses `fmt.Errorf` or stdlib `errors.New`.

### 3. Bug in `observability/metrics/otelgrpc/metrics_test.go`

Subtests reference parent test variable `T` instead of subtest variable `t` in ~11 testify assertions. Assertion failures would be attributed to the wrong test scope.

### 4. `init()` functions that panic

- `observability/logging/zerolog/zerolog_logger.go` — panics if `time.LoadLocation("America/Chicago")` fails (also a hardcoded timezone)
- `random/random.go` — panics if crypto/rand init fails
- `observability/logging/zap/zap_logger.go` — panics if zap logger creation fails

### 5. Missing `wire.go` and `do.go` in 5 config packages

`cache/config`, `eventstream/config`, `cryptography/encryption/config`, `observability/metrics/otelgrpc/config`, `search/text/config` — all missing both files, breaking the DI pattern.

## Medium

### 6. OpenTelemetry instrumentation gaps

`messagequeue` (sqs, redis, pubsub) and `eventstream` (sse, websocket) have no tracing or metrics, despite CLAUDE.md stating "OpenTelemetry throughout" for messaging.

### 7. `context.Background()` in place of parent context

- `messagequeue/redis/consumer.go:53`
- `messagequeue/sqs/consumer.go:56`
- `messagequeue/pubsub/consumer.go:49`
- `circuitbreaking/circuitbreaking.go:109`
- `server/grpc/server.go:119`

These ignore cancellation signals from callers.

### 8. Incomplete zap logger `SetLevel()`

`observability/logging/zap/zap_logger.go:48-67` — computes a level but never applies it. Dead code with `_ = lvl`.

### 9. Duplicated logger implementations

All 4 logger backends (slog, otelgrpc/slog, zap, zerolog) duplicate nearly identical `WithName`, `WithValue`, `WithValues`, `WithError`, `WithSpan`, `WithRequest`, `WithResponse`, and `attachRequestToLog` implementations.

### 10. `database/TODO.md` says package is due for deletion/move

Either do it or remove the TODO — stale intent is confusing.

### 11. 41 packages have no tests

Many are mock/noop packages (acceptable), but notable gaps include: `errors/`, `errors/grpc/`, `errors/http/`, `internalerrors/`, `capitalism/`, `analytics/`, `version/`, `testutils/`.

### 12. stdlib `log` usage instead of platform logger

- `circuitbreaking/circuitbreaking.go:49` — `log.Println()`
- `messagequeue/redis/test_helpers.go:46` — `log.Printf()`
