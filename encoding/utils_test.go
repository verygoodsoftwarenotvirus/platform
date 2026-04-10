package encoding

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(T *testing.T) {
	T.Parallel()

	T.Run("with nil content type", func(t *testing.T) {
		t.Parallel()

		var dest example
		assert.NoError(t, Decode([]byte(`{"name":"test"}`), nil, &dest))
	})

	T.Run("with explicit content type", func(t *testing.T) {
		t.Parallel()

		var dest example
		assert.NoError(t, Decode([]byte(`<example><name>test</name></example>`), ContentTypeXML, &dest))
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		var dest example
		assert.Error(t, Decode([]byte(`{invalid`), nil, &dest))
	})
}

func TestMustEncode(T *testing.T) {
	T.Parallel()

	T.Run("with nil content type", func(t *testing.T) {
		t.Parallel()

		result := MustEncode(&example{Name: t.Name()}, nil)
		assert.NotEmpty(t, result)
	})

	T.Run("with explicit content type", func(t *testing.T) {
		t.Parallel()

		result := MustEncode(&example{Name: t.Name()}, ContentTypeXML)
		assert.NotEmpty(t, result)
	})

	T.Run("panics with un-encodable data", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.NotNil(t, recover())
		}()

		MustEncode(&broken{Name: json.Number(t.Name())}, nil)
	})
}

func TestMustDecode(T *testing.T) {
	T.Parallel()

	T.Run("with nil content type", func(t *testing.T) {
		t.Parallel()

		var dest example
		MustDecode([]byte(`{"name":"test"}`), nil, &dest)
	})

	T.Run("with explicit content type", func(t *testing.T) {
		t.Parallel()

		var dest example
		MustDecode([]byte(`<example><name>test</name></example>`), ContentTypeXML, &dest)
	})

	T.Run("panics with invalid data", func(t *testing.T) {
		t.Parallel()

		defer func() {
			assert.NotNil(t, recover())
		}()

		var dest example
		MustDecode([]byte(`{invalid`), nil, &dest)
	})
}

func TestMustEncodeJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		result := MustEncodeJSON(&example{Name: t.Name()})
		assert.NotEmpty(t, result)
	})
}

func TestDecodeJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var dest example
		assert.NoError(t, DecodeJSON([]byte(`{"name":"test"}`), &dest))
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		var dest example
		assert.Error(t, DecodeJSON([]byte(`{invalid`), &dest))
	})
}

func TestMustDecodeJSON(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var dest example
		MustDecodeJSON([]byte(`{"name":"test"}`), &dest)
	})
}

func TestMustJSONIntoReader(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		reader := MustJSONIntoReader(&example{Name: t.Name()})
		require.NotNil(t, reader)

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})
}
