package grpc

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

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
		_, err := NewGRPCServer(cfg, logging.NewNoopLogger(), nil, nil, nil)
		require.NoError(t, err)
	})
}
