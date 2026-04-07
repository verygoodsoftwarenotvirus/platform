package chi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLoggingMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		tracer := tracing.NewTracerForTest("")
		middleware := buildLoggingMiddleware(logging.NewNoopLogger(), tracer, false)

		assert.NotNil(t, middleware)

		hf := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {})

		req, res := httptest.NewRequestWithContext(ctx, http.MethodPost, "/nil", http.NoBody), httptest.NewRecorder()

		middleware(hf).ServeHTTP(res, req)
	})
}
