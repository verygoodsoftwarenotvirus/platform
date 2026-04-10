package encoding

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

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

		assert.NotNil(t, buildContentType("test"))
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

	T.Run("with TOML", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeTOML, contentTypeFromString(contentTypeTOML))
	})

	T.Run("with YAML", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeYAML, contentTypeFromString(contentTypeYAML))
	})

	T.Run("with Emoji", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeEmoji, contentTypeFromString(contentTypeEmoji))
	})

	T.Run("with unknown defaults to JSON", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeJSON, contentTypeFromString("unknown"))
	})
}
