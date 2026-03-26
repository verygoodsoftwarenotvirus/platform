package chi

import (
	"context"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/identifiers"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := identifiers.New()
		ctx := context.WithValue(t.Context(), chimiddleware.RequestIDKey, expected)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", http.NoBody)
		require.NoError(t, err)

		actual := RequestIDFunc(req)
		assert.Equal(t, expected, actual)
	})
}
