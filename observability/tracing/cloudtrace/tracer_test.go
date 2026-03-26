package cloudtrace

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
)

func Test_tracingErrorHandler_Handle(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		errorHandler{logger: logging.NewNoopLogger()}.Handle(errors.New("blah"))
	})
}
