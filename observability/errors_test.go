package observability

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		descriptionFmt, descriptionArgs := "things and %s", "stuff"
		err := errors.New("blah")

		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareError(err, span, descriptionFmt, descriptionArgs))
	})
}

func TestAcknowledgeError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		descriptionFmt, descriptionArgs := "things and %s", "stuff"
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		AcknowledgeError(err, logger, span, descriptionFmt, descriptionArgs)
	})
}
