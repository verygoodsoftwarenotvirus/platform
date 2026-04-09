package observability

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestPrepareAndLogError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogError(err, logger, span, "things and %s", "stuff"))
	})

	T.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.NoError(t, PrepareAndLogError(nil, logger, span, "things and %s", "stuff"))
	})

	T.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		err := errors.New("blah")
		logger := logging.NewNoopLogger()

		assert.Error(t, PrepareAndLogError(err, logger, nil, "things and %s", "stuff"))
	})

	T.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogError(err, nil, span, "things and %s", "stuff"))
	})

	T.Run("with empty description", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogError(err, logger, span, ""))
	})
}

func TestPrepareError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareError(err, span, "things and %s", "stuff"))
	})

	T.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		_, span := tracing.StartSpan(ctx)

		assert.NoError(t, PrepareError(nil, span, "things and %s", "stuff"))
	})

	T.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		err := errors.New("blah")

		assert.Error(t, PrepareError(err, nil, "things and %s", "stuff"))
	})

	T.Run("with empty description", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		_, span := tracing.StartSpan(ctx)

		actual := PrepareError(err, span, "")
		assert.Error(t, actual)
		assert.Equal(t, err, actual)
	})
}

func TestAcknowledgeError(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		AcknowledgeError(err, logger, span, "things and %s", "stuff")
	})

	T.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		err := errors.New("blah")
		logger := logging.NewNoopLogger()

		AcknowledgeError(err, logger, nil, "things and %s", "stuff")
	})

	T.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		_, span := tracing.StartSpan(ctx)

		AcknowledgeError(err, nil, span, "things and %s", "stuff")
	})

	T.Run("with empty description", func(t *testing.T) {
		t.Parallel()

		err := errors.New("blah")
		logger := logging.NewNoopLogger()

		AcknowledgeError(err, logger, nil, "")
	})
}

func TestPrepareAndLogGRPCStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "things and %s", "stuff"))
	})

	T.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.NoError(t, PrepareAndLogGRPCStatus(nil, logger, span, codes.Internal, "things and %s", "stuff"))
	})

	T.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		err := errors.New("blah")
		logger := logging.NewNoopLogger()

		assert.Error(t, PrepareAndLogGRPCStatus(err, logger, nil, codes.Internal, "things and %s", "stuff"))
	})

	T.Run("with nil logger", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogGRPCStatus(err, nil, span, codes.Internal, "things and %s", "stuff"))
	})

	T.Run("with empty description", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		err := errors.New("blah")
		logger := logging.NewNoopLogger()
		_, span := tracing.StartSpan(ctx)

		assert.Error(t, PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, ""))
	})
}
