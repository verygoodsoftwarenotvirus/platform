package http

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v5/panicking"
	"github.com/verygoodsoftwarenotvirus/platform/v5/routing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type mockTracerProvider struct {
	noop.TracerProvider
	mock.Mock
}

func (m *mockTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return noop.NewTracerProvider().Tracer(name, opts...)
}

func (m *mockTracerProvider) ForceFlush(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// stubRouter satisfies routing.Router for testing Serve().
type stubRouter struct{}

func (stubRouter) Routes() []*routing.Route                            { return nil }
func (stubRouter) Handler() http.Handler                               { return http.NewServeMux() }
func (stubRouter) Handle(string, http.Handler)                         {}
func (stubRouter) HandleFunc(string, http.HandlerFunc)                 {}
func (stubRouter) WithMiddleware(...routing.Middleware) routing.Router { return stubRouter{} }
func (stubRouter) Route(string, func(r routing.Router)) routing.Router { return stubRouter{} }
func (stubRouter) Connect(string, http.HandlerFunc)                    {}
func (stubRouter) Delete(string, http.HandlerFunc)                     {}
func (stubRouter) Get(string, http.HandlerFunc)                        {}
func (stubRouter) Head(string, http.HandlerFunc)                       {}
func (stubRouter) Options(string, http.HandlerFunc)                    {}
func (stubRouter) Patch(string, http.HandlerFunc)                      {}
func (stubRouter) Post(string, http.HandlerFunc)                       {}
func (stubRouter) Put(string, http.HandlerFunc)                        {}
func (stubRouter) Trace(string, http.HandlerFunc)                      {}
func (stubRouter) AddRoute(string, string, http.HandlerFunc, ...routing.Middleware) error {
	return nil
}

func generateTestTLSCerts(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	dir := t.TempDir()

	certPath := filepath.Join(dir, "cert.pem")
	certOut, err := os.Create(certPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	require.NoError(t, certOut.Close())

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPath := filepath.Join(dir, "key.pem")
	keyOut, err := os.Create(keyPath)
	require.NoError(t, err)
	require.NoError(t, pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	require.NoError(t, keyOut.Close())

	return certPath, keyPath
}

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

	T.Run("logs error when ForceFlush fails", func(t *testing.T) {
		t.Parallel()

		mtp := &mockTracerProvider{}
		mtp.On("ForceFlush", mock.Anything).Return(errors.New("flush failed"))

		s, err := ProvideHTTPServer(Config{Port: 0}, logging.NewNoopLogger(), nil, mtp, "")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		assert.NoError(t, s.Shutdown(ctx))

		mock.AssertExpectationsForObjects(t, mtp)
	})
}

func TestServer_Serve(T *testing.T) {
	T.Parallel()

	T.Run("serves HTTP and shuts down cleanly", func(t *testing.T) {
		t.Parallel()

		srv := &server{
			logger:         logging.NewNoopLogger(),
			router:         stubRouter{},
			panicker:       panicking.NewProductionPanicker(),
			httpServer:     provideStdLibHTTPServer(0),
			tracerProvider: tracing.NewNoopTracerProvider(),
			config:         Config{},
		}

		done := make(chan struct{})
		go func() {
			srv.Serve()
			close(done)
		}()

		// Give the server time to start listening.
		time.Sleep(50 * time.Millisecond)
		require.NoError(t, srv.httpServer.Close())
		<-done
	})

	T.Run("serves HTTPS and shuts down cleanly", func(t *testing.T) {
		t.Parallel()

		certFile, keyFile := generateTestTLSCerts(t)

		srv := &server{
			logger:         logging.NewNoopLogger(),
			router:         stubRouter{},
			panicker:       panicking.NewProductionPanicker(),
			httpServer:     provideStdLibHTTPServer(0),
			tracerProvider: tracing.NewNoopTracerProvider(),
			config: Config{
				SSLCertificateFile:    certFile,
				SSLCertificateKeyFile: keyFile,
			},
		}

		done := make(chan struct{})
		go func() {
			srv.Serve()
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		require.NoError(t, srv.httpServer.Close())
		<-done
	})

	T.Run("logs error for HTTPS with invalid cert files", func(t *testing.T) {
		t.Parallel()

		srv := &server{
			logger:         logging.NewNoopLogger(),
			router:         stubRouter{},
			panicker:       panicking.NewProductionPanicker(),
			httpServer:     provideStdLibHTTPServer(0),
			tracerProvider: tracing.NewNoopTracerProvider(),
			config: Config{
				SSLCertificateFile:    "/nonexistent/cert.pem",
				SSLCertificateKeyFile: "/nonexistent/key.pem",
			},
		}

		// ListenAndServeTLS fails immediately with invalid cert paths.
		srv.Serve()
	})

	T.Run("logs error for HTTP listen failure", func(t *testing.T) {
		t.Parallel()

		// Occupy a port so ListenAndServe fails with "address already in use".
		lis, err := new(net.ListenConfig).Listen(t.Context(), "tcp", ":0")
		require.NoError(t, err)
		defer lis.Close()

		port := lis.Addr().(*net.TCPAddr).Port

		httpSrv := provideStdLibHTTPServer(uint16(port))

		srv := &server{
			logger:         logging.NewNoopLogger(),
			router:         stubRouter{},
			panicker:       panicking.NewProductionPanicker(),
			httpServer:     httpSrv,
			tracerProvider: tracing.NewNoopTracerProvider(),
			config:         Config{},
		}

		srv.Serve()
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
