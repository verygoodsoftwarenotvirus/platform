package encoding

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func Test_clientEncoder_ContentType(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		assert.NotEmpty(t, e.ContentType())
	})
}

func Test_buildContentType(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		//
	})
}

func TestContentTypeToString(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, ContentTypeToString(ContentTypeJSON))
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, ContentTypeToString(ContentTypeXML))
	})

	T.Run("with Emoji", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, ContentTypeToString(ContentTypeEmoji))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, ContentTypeToString(nil))
	})
}

func Test_contentTypeFromString(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeJSON, contentTypeFromString(contentTypeJSON))
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeXML, contentTypeFromString(contentTypeXML))
	})
}
