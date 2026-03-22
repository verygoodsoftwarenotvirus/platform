package encoding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v2/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProvideClientEncoder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON))
	})
}

func Test_clientEncoder_Unmarshal(T *testing.T) {
	T.Parallel()

	testCases := map[string]struct {
		contentType ContentType
		expected    string
	}{
		"json": {
			contentType: ContentTypeJSON,
			expected:    `{"name": "name"}`,
		},
		"xml": {
			contentType: ContentTypeXML,
			expected:    `<example><name>name</name></example>`,
		},
		"toml": {
			contentType: ContentTypeTOML,
			expected:    `name = "name"`,
		},
		"yaml": {
			contentType: ContentTypeYAML,
			expected:    `name: "name"`,
		},
		"emoji": {
			contentType: ContentTypeEmoji,
			expected:    "🍃🧁🌆🙍☔🌾🐯🦮💆🚂🚕🏏🧔✊🀄🏏☔🌊🥈🐾👥♓🙌🀄🀄🍧🦖📓♿😱🦨🐶🀄☕\n",
		},
	}

	for name, tc := range testCases {
		T.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), tc.contentType)

			expected := &example{Name: "name"}
			actual := &example{}

			assert.NoError(t, e.Unmarshal(ctx, []byte(tc.expected), &actual))
			assert.Equal(t, expected, actual)
		})
	}

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		actual := &example{}

		assert.Error(t, e.Unmarshal(ctx, []byte(`{"name"   `), &actual))
		assert.Empty(t, actual.Name)
	})
}

func Test_clientEncoder_Encode(T *testing.T) {
	T.Parallel()

	for _, ct := range ContentTypes {
		T.Run(ContentTypeToString(ct), func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ct)

			res := httptest.NewRecorder()

			assert.NoError(t, e.Encode(ctx, res, &example{Name: t.Name()}))
		})
	}

	for _, ct := range ContentTypes {
		T.Run(fmt.Sprintf("%s handles io.Writer errors", ContentTypeToString(ct)), func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ct)

			mw := &mockWriter{}
			mw.On(reflection.GetMethodName(mw.Write), mock.Anything).Return(0, errors.New("blah"))

			assert.Error(t, e.Encode(ctx, mw, &example{Name: t.Name()}))
		})
	}

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		assert.Error(t, e.Encode(ctx, nil, &broken{Name: json.Number(t.Name())}))
	})
}

func Test_clientEncoder_EncodeReader(T *testing.T) {
	T.Parallel()

	for _, ct := range ContentTypes {
		T.Run(ContentTypeToString(ct), func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ct)

			actual, err := e.EncodeReader(ctx, &example{Name: t.Name()})
			assert.NoError(t, err)
			assert.NotNil(t, actual)
		})
	}

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		e := ProvideClientEncoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		actual, err := e.EncodeReader(ctx, &broken{Name: json.Number(t.Name())})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
