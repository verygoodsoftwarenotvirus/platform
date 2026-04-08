package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideHTTPServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x, err := ProvideHTTPServer(
			Config{
				SSLCertificateFile:    "",
				SSLCertificateKeyFile: "",
				StartupDeadline:       0,
				Port:                  0,
				Debug:                 false,
			},
			nil,
			nil,
			nil,
			"",
		)

		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with custom service name", func(t *testing.T) {
		t.Parallel()

		x, err := ProvideHTTPServer(
			Config{Port: 8080},
			logging.NewNoopLogger(),
			nil,
			nil,
			"custom_service",
		)

		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with empty service name uses default", func(t *testing.T) {
		t.Parallel()

		x, err := ProvideHTTPServer(
			Config{Port: 8080},
			logging.NewNoopLogger(),
			nil,
			nil,
			"",
		)

		assert.NotNil(t, x)
		assert.NoError(t, err)
	})

	T.Run("with SSL config", func(t *testing.T) {
		t.Parallel()

		x, err := ProvideHTTPServer(
			Config{
				SSLCertificateFile:    "/some/cert.pem",
				SSLCertificateKeyFile: "/some/key.pem",
				Port:                  8443,
			},
			logging.NewNoopLogger(),
			nil,
			nil,
			"",
		)

		assert.NotNil(t, x)
		assert.NoError(t, err)
	})
}

func TestServer_Router(T *testing.T) {
	T.Parallel()

	T.Run("returns the router", func(t *testing.T) {
		t.Parallel()

		s, err := ProvideHTTPServer(Config{Port: 0}, nil, nil, nil, "")
		require.NoError(t, err)

		// Router returns nil when nil was passed in
		assert.Nil(t, s.Router())
	})
}

func TestServer_Shutdown(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s, err := ProvideHTTPServer(Config{Port: 0}, logging.NewNoopLogger(), nil, nil, "")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		assert.NoError(t, s.Shutdown(ctx))
	})
}

func Test_skipNoisePaths(T *testing.T) {
	T.Parallel()

	T.Run("ops paths are filtered out", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/_ops_/health", http.NoBody)
		assert.False(t, skipNoisePaths(req))
	})

	T.Run("apple app site association path is filtered out", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, appleAppSiteAssociationPath, http.NoBody)
		assert.False(t, skipNoisePaths(req))
	})

	T.Run("normal paths are not filtered", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/things", http.NoBody)
		assert.True(t, skipNoisePaths(req))
	})

	T.Run("root path is not filtered", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		assert.True(t, skipNoisePaths(req))
	})
}

func Test_provideStdLibHTTPServer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		srv := provideStdLibHTTPServer(8080)

		assert.NotNil(t, srv)
		assert.Equal(t, ":8080", srv.Addr)
		assert.Equal(t, readTimeout, srv.ReadTimeout)
		assert.Equal(t, writeTimeout, srv.WriteTimeout)
		assert.Equal(t, idleTimeout, srv.IdleTimeout)
		assert.NotNil(t, srv.TLSConfig)
		assert.Equal(t, uint16(tls.VersionTLS12), srv.TLSConfig.MinVersion)
	})

	T.Run("with zero port", func(t *testing.T) {
		t.Parallel()

		srv := provideStdLibHTTPServer(0)

		assert.NotNil(t, srv)
		assert.Equal(t, ":0", srv.Addr)
	})
}
