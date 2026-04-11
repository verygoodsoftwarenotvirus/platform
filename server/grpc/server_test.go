package grpc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
)

type mockTracerProvider struct {
	noop.TracerProvider
	forceFlushFunc  func(ctx context.Context) error
	forceFlushCalls int
}

func (m *mockTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return noop.NewTracerProvider().Tracer(name, opts...)
}

func (m *mockTracerProvider) ForceFlush(ctx context.Context) error {
	m.forceFlushCalls++
	if m.forceFlushFunc == nil {
		return nil
	}
	return m.forceFlushFunc(ctx)
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

func TestNewGRPCServer(T *testing.T) {
	T.Parallel()

	T.Run("returns error with nil config", func(t *testing.T) {
		t.Parallel()

		server, err := NewGRPCServer(nil, nil, nil, nil, nil)

		assert.Nil(t, server)
		assert.Error(t, err)
	})

	T.Run("succeeds with valid config", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, nil, nil, nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, server)
	})

	T.Run("succeeds with registration functions", func(t *testing.T) {
		t.Parallel()

		called := false
		rf := func(s *grpc.Server) {
			called = true
		}

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil, rf)

		require.NoError(t, err)
		assert.NotNil(t, server)
		assert.True(t, called)
	})

	T.Run("returns error with invalid TLS files", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Port:                  0,
			HTTPSCertificateFile:  "/nonexistent/cert.pem",
			TLSCertificateKeyFile: "/nonexistent/key.pem",
		}

		server, err := NewGRPCServer(cfg, nil, nil, nil, nil)

		assert.Nil(t, server)
		assert.Error(t, err)
	})

	T.Run("succeeds with valid TLS files", func(t *testing.T) {
		t.Parallel()

		certFile, keyFile := generateTestTLSCerts(t)

		cfg := &Config{
			Port:                  0,
			HTTPSCertificateFile:  certFile,
			TLSCertificateKeyFile: keyFile,
		}

		server, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil)

		require.NoError(t, err)
		assert.NotNil(t, server)
	})
}

func TestLoggingInterceptor(T *testing.T) {
	T.Parallel()

	T.Run("returns interceptor that calls handler", func(t *testing.T) {
		t.Parallel()

		interceptor := LoggingInterceptor(nil)
		assert.NotNil(t, interceptor)

		handlerCalled := false
		handler := func(ctx context.Context, req any) (any, error) {
			handlerCalled = true
			return "result", nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		result, err := interceptor(context.Background(), "request", info, handler)

		assert.NoError(t, err)
		assert.Equal(t, "result", result)
		assert.True(t, handlerCalled)
	})

	T.Run("logs error when handler fails", func(t *testing.T) {
		t.Parallel()

		interceptor := LoggingInterceptor(logging.NewNoopLogger())
		assert.NotNil(t, interceptor)

		expectedErr := assert.AnError
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, expectedErr
		}

		info := &grpc.UnaryServerInfo{FullMethod: "/test/Method"}
		result, err := interceptor(context.Background(), "request", info, handler)

		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})
}

func TestServer_Shutdown(T *testing.T) {
	T.Parallel()

	T.Run("shuts down gracefully", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, nil, nil, nil, nil)
		require.NoError(t, err)

		server.Shutdown(context.Background())
	})

	T.Run("shuts down with logger", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil)
		require.NoError(t, err)

		server.Shutdown(context.Background())
	})

	T.Run("logs error when ForceFlush fails", func(t *testing.T) {
		t.Parallel()

		mtp := &mockTracerProvider{
			forceFlushFunc: func(_ context.Context) error { return errors.New("flush failed") },
		}

		cfg := &Config{Port: 0}
		srv, err := NewGRPCServer(cfg, logging.NewNoopLogger(), mtp, nil, nil)
		require.NoError(t, err)

		srv.Shutdown(context.Background())

		assert.Equal(t, 1, mtp.forceFlushCalls)
	})
}

func TestNewGRPCServer_withInterceptors(T *testing.T) {
	T.Parallel()

	T.Run("with unary interceptors", func(t *testing.T) {
		t.Parallel()

		unaryInterceptor := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			return handler(ctx, req)
		}

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, []grpc.UnaryServerInterceptor{unaryInterceptor}, nil)

		require.NoError(t, err)
		assert.NotNil(t, server)
	})

	T.Run("with stream interceptors", func(t *testing.T) {
		t.Parallel()

		streamInterceptor := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, []grpc.StreamServerInterceptor{streamInterceptor})

		require.NoError(t, err)
		assert.NotNil(t, server)
	})

	T.Run("with multiple registration functions", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		rf1 := func(s *grpc.Server) { callCount++ }
		rf2 := func(s *grpc.Server) { callCount++ }

		cfg := &Config{Port: 0}
		server, err := NewGRPCServer(cfg, nil, nil, nil, nil, rf1, rf2)

		require.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, 2, callCount)
	})
}

func TestServer_Serve(T *testing.T) {
	T.Parallel()

	T.Run("serves and can be stopped", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{Port: 0}
		srv, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(t.Context())

		done := make(chan struct{})
		go func() {
			srv.Serve(ctx)
			close(done)
		}()

		// Give the server a moment to start listening, then stop it.
		time.Sleep(50 * time.Millisecond)
		srv.grpcServer.GracefulStop()
		cancel()
		<-done
	})

	T.Run("returns when listen fails", func(t *testing.T) {
		t.Parallel()

		// Occupy a port so the server's Listen call fails with "address already in use".
		lis, err := new(net.ListenConfig).Listen(t.Context(), "tcp", ":0")
		require.NoError(t, err)
		defer lis.Close()

		port := lis.Addr().(*net.TCPAddr).Port

		cfg := &Config{Port: uint16(port)}
		srv, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil)
		require.NoError(t, err)

		// Should return immediately because the port is already in use.
		srv.Serve(t.Context())
	})
}
