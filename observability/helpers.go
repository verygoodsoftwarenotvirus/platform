package observability

import (
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
)

func ObserveValues(values map[string]any, span tracing.Span, logger logging.Logger) logging.Logger {
	for k, v := range values {
		if span != nil {
			tracing.AttachToSpan(span, k, v)
		}

		if logger != nil {
			logger = logger.WithValue(k, v)
			if span != nil {
				logger = logger.WithSpan(span)
			}
		}
	}

	return logger
}
