package encoding

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type example struct {
	Name string `json:"name" xml:"name"`
}

type broken struct {
	Name json.Number `json:"name" xml:"name"`
}

func init() {
	gob.Register(&example{})
	gob.Register(&broken{})
}

func TestServerEncoderDecoder_encodeResponse(T *testing.T) {
	T.Parallel()

	testCases := map[string]struct {
		contentType      ContentType
		expectedResponse string
	}{
		"json": {
			contentType:      ContentTypeJSON,
			expectedResponse: `{"name":"name"}` + "\n",
		},
		"xml": {
			contentType:      ContentTypeXML,
			expectedResponse: "<example><name>name</name></example>",
		},
		"toml": {
			contentType:      ContentTypeTOML,
			expectedResponse: `Name = "name"` + "\n",
		},
		"yaml": {
			contentType:      ContentTypeYAML,
			expectedResponse: "name: name\n",
		},
	}

	for testName, tc := range testCases {
		T.Run(testName, func(t *testing.T) {
			t.Parallel()

			ex := &example{Name: "name"}
			encoderDecoder, ok := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), tc.contentType).(*serverEncoderDecoder)
			require.True(t, ok)

			ctx := t.Context()
			res := httptest.NewRecorder()
			res.Header().Set(ContentTypeHeaderKey, ContentTypeToString(tc.contentType))

			encoderDecoder.encodeResponse(ctx, res, ex, http.StatusOK)
			actual := res.Body.String()
			assert.Equal(t, tc.expectedResponse, actual)
		})
	}

	T.Run("emoji", func(t *testing.T) {
		t.Parallel()

		ex := &example{Name: "name"}
		encoderDecoder, ok := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeEmoji).(*serverEncoderDecoder)
		require.True(t, ok)

		ctx := t.Context()
		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeaderKey, ContentTypeToString(ContentTypeEmoji))

		encoderDecoder.encodeResponse(ctx, res, ex, http.StatusOK)
		actual := res.Body.String()
		assert.NotEmpty(t, actual)
	})

	T.Run("defaults to JSON", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		encoderDecoder, ok := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON).(*serverEncoderDecoder)
		require.True(t, ok)

		ctx := t.Context()
		res := httptest.NewRecorder()

		encoderDecoder.encodeResponse(ctx, res, ex, http.StatusOK)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})

	T.Run("with broken structure", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &broken{Name: json.Number(expectation)}
		encoderDecoder, ok := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON).(*serverEncoderDecoder)
		require.True(t, ok)

		ctx := t.Context()
		res := httptest.NewRecorder()

		encoderDecoder.encodeResponse(ctx, res, ex, http.StatusOK)
		assert.Empty(t, res.Body.String())
	})
}

func TestServerEncoderDecoder_MustEncodeJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		expected := `{"name":"TestServerEncoderDecoder_MustEncodeJSON/standard"}
`
		actual := string(encoderDecoder.MustEncodeJSON(ctx, &example{Name: t.Name()}))

		assert.Equal(t, expected, actual)
	})

	T.Run("with panic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		defer func() {
			assert.NotNil(t, recover())
		}()

		encoderDecoder.MustEncodeJSON(ctx, &broken{Name: json.Number(t.Name())})
	})
}

func TestServerEncoderDecoder_MustEncode(T *testing.T) {
	T.Parallel()

	testCases := map[string]struct {
		contentType ContentType
		expected    string
	}{
		"json": {
			contentType: ContentTypeJSON,
			expected:    `{"name":"TestServerEncoderDecoder_MustEncode/json"}` + "\n",
		},
		"xml": {
			contentType: ContentTypeXML,
			expected:    "<example><name>TestServerEncoderDecoder_MustEncode/xml</name></example>",
		},
		"toml": {
			contentType: ContentTypeTOML,
			expected:    "Name = \"TestServerEncoderDecoder_MustEncode/toml\"\n",
		},
		"yaml": {
			contentType: ContentTypeYAML,
			expected:    "name: TestServerEncoderDecoder_MustEncode/yaml\n",
		},
	}

	for name, tc := range testCases {
		T.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), tc.contentType)

			actual := string(encoderDecoder.MustEncode(ctx, &example{Name: t.Name()}))

			assert.Equal(t, tc.expected, actual)
		})
	}

	T.Run("emoji", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeEmoji)

		actual := string(encoderDecoder.MustEncode(ctx, &example{Name: t.Name()}))
		assert.NotEmpty(t, actual)
	})

	T.Run("with broken struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		encoderDecoder, ok := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON).(*serverEncoderDecoder)
		require.True(t, ok)

		defer func() {
			assert.NotNil(t, recover())
		}()

		encoderDecoder.MustEncode(ctx, &broken{Name: json.Number(t.Name())})
	})
}

func TestServerEncoderDecoder_EncodeResponseWithStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), ContentTypeJSON)

		ctx := t.Context()
		res := httptest.NewRecorder()

		expected := 666
		encoderDecoder.EncodeResponseWithStatus(ctx, res, ex, expected)

		assert.Equal(t, expected, res.Code, "expected code to be %d, but got %d", expected, res.Code)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})
}

func TestServerEncoderDecoder_DecodeRequest(T *testing.T) {
	T.Parallel()

	testCases := map[string]struct {
		contentType ContentType
		marshaller  func(v any) ([]byte, error)
		expected    string
	}{
		"json": {
			contentType: ContentTypeJSON,
			expected:    `{"name":"name"}`,
			marshaller:  json.Marshal,
		},
		"xml": {
			contentType: ContentTypeXML,
			expected:    `<example><name>name</name></example>`,
			marshaller:  xml.Marshal,
		},
		"toml": {
			contentType: ContentTypeTOML,
			expected:    `<example><name>name</name></example>`,
			marshaller:  tomlMarshalFunc,
		},
		"yaml": {
			contentType: ContentTypeYAML,
			expected:    `<example><name>name</name></example>`,
			marshaller:  yaml.Marshal,
		},
		"emoji": {
			contentType: ContentTypeEmoji,
			expected:    `<example><name>name</name></example>`,
			marshaller:  marshalEmoji,
		},
	}

	e := &example{Name: "name"}

	for name, tc := range testCases {
		T.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), tc.contentType)

			bs, err := tc.marshaller(e)
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"https://whatever.whocares.gov",
				bytes.NewReader(bs),
			)
			require.NoError(t, err)
			req.Header.Set(ContentTypeHeaderKey, ContentTypeToString(tc.contentType))

			var x example
			assert.NoError(t, encoderDecoder.DecodeRequest(ctx, req, &x))
			assert.Equal(t, x.Name, e.Name)
		})
	}
}

func Test_serverEncoderDecoder_DecodeBytes(T *testing.T) {
	T.Parallel()

	goodDataTestCases := map[string]struct {
		contentType ContentType
		data        []byte
	}{
		"json": {
			data:        []byte(`{"name":"name"}`),
			contentType: ContentTypeJSON,
		},
		"xml": {
			data:        []byte(`<example><name>name</name></example>`),
			contentType: ContentTypeXML,
		},
		"toml": {
			data:        []byte(`name = "name"`),
			contentType: ContentTypeTOML,
		},
		"yaml": {
			data:        []byte(`name: "name"`),
			contentType: ContentTypeYAML,
		},
		"emoji": {
			data:        []byte("🍃🧁🌆🙍☔🌾🐯🦮💆🚂🚕🏏🧔✊🀄🏏☔🌊🥈🐾👥♓🙌🀄🀄🍧🦖📓♿😱🦨🐶🀄☕\n"),
			contentType: ContentTypeEmoji,
		},
	}
	goodDataExpectation := &example{Name: "name"}

	for name, tc := range goodDataTestCases {
		T.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			encoderDecoder := ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), tc.contentType)

			var dest *example
			assert.NoError(t, encoderDecoder.DecodeBytes(ctx, tc.data, &dest))

			assert.Equal(t, goodDataExpectation, dest)
		})
	}
}
