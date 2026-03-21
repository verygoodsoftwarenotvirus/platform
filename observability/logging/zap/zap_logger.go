package zap

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger is our log wrapper.
type zapLogger struct {
	requestIDFunc logging.RequestIDFunc
	logger        *zap.Logger
}

// NewZapLogger builds a new zapLogger.
func NewZapLogger(lvl logging.Level) logging.Logger {
	switch lvl {
	case logging.DebugLevel:
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}

		return &zapLogger{logger: l.WithOptions(zap.AddCallerSkip(1))}
	default:
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}

		return &zapLogger{logger: l.WithOptions(zap.AddCallerSkip(1))}
	}
}

// WithName is our obligatory contract fulfillment function.
func (l *zapLogger) WithName(name string) logging.Logger {
	l2 := l.logger.With(zap.String(logging.LoggerNameKey, name))
	return &zapLogger{logger: l2}
}

// SetLevel sets the log level for our zap logger.
func (l *zapLogger) SetLevel(level logging.Level) {
	var lvl zapcore.Level

	switch level {
	case logging.InfoLevel:
		lvl = zap.InfoLevel
	case logging.DebugLevel:
		l.logger.WithOptions(zap.WithCaller(true))
		lvl = zap.DebugLevel
	case logging.WarnLevel:
		lvl = zap.WarnLevel
	case logging.ErrorLevel:
		lvl = zap.ErrorLevel
	default:
		lvl = zap.InfoLevel
	}

	_ = lvl
	// there isn't really a way to set the level of a zap logger, but this doesn't really seem to get called anyway lol
}

// SetRequestIDFunc sets the request ID retrieval function.
func (l *zapLogger) SetRequestIDFunc(f logging.RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *zapLogger) Info(input string) {
	l.logger.Info(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *zapLogger) Debug(input string) {
	l.logger.Debug(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *zapLogger) Error(whatWasHappeningWhenErrorOccurred string, err error) {
	if err != nil {
		l.logger.Error(fmt.Sprintf("error %s: %s", whatWasHappeningWhenErrorOccurred, err.Error()))
	}
}

// Clone satisfies our contract for the logging.Logger WithValue method.
func (l *zapLogger) Clone() logging.Logger {
	l2 := l.logger.With()
	return &zapLogger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *zapLogger) WithValue(key string, value any) logging.Logger {
	l2 := l.logger.With(zap.Any(key, value))
	return &zapLogger{logger: l2}
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *zapLogger) WithValues(values map[string]any) logging.Logger {
	var l2 = l.logger.With()

	for key, val := range values {
		l2 = l2.With(zap.Any(key, val))
	}

	return &zapLogger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *zapLogger) WithError(err error) logging.Logger {
	l2 := l.logger.With(zap.Error(err))
	return &zapLogger{logger: l2}
}

// WithSpan satisfies our contract for the logging.Logger WithSpan method.
func (l *zapLogger) WithSpan(span trace.Span) logging.Logger {
	spanCtx := span.SpanContext()
	spanID := spanCtx.SpanID().String()
	traceID := spanCtx.TraceID().String()

	l2 := l.logger.With(zap.String(keys.SpanIDKey, spanID), zap.String(keys.TraceIDKey, traceID))

	return &zapLogger{logger: l2}
}

func (l *zapLogger) attachRequestToLog(req *http.Request) *zap.Logger {
	if req != nil {
		l2 := l.logger.With(zap.String("method", req.Method))

		if req.URL != nil {
			l2 = l2.With(zap.String("path", req.URL.Path))
			if req.URL.RawQuery != "" {
				l2 = l2.With(zap.String(keys.URLQueryKey, req.URL.RawQuery))
			}
		}

		if l.requestIDFunc != nil {
			if reqID := l.requestIDFunc(req); reqID != "" {
				l2 = l2.With(zap.String("request.id", reqID))
			}
		}

		return l2
	}

	return l.logger
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *zapLogger) WithRequest(req *http.Request) logging.Logger {
	return &zapLogger{logger: l.attachRequestToLog(req)}
}

// WithResponse satisfies our contract for the logging.Logger WithResponse method.
func (l *zapLogger) WithResponse(res *http.Response) logging.Logger {
	l2 := l.logger.With()
	if res != nil {
		l2 = l.attachRequestToLog(res.Request).With(zap.Int(keys.ResponseStatusKey, res.StatusCode))
	}

	return &zapLogger{logger: l2}
}
