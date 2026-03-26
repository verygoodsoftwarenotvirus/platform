package grpc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v4/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestDecodeErrorFromStatus(T *testing.T) {
	T.Parallel()

	T.Run("nil error returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, DecodeErrorFromStatus(context.Background(), nil))
	})

	T.Run("non-status error returned as-is", func(t *testing.T) {
		t.Parallel()
		original := errors.New("plain error")
		result := DecodeErrorFromStatus(context.Background(), original)
		assert.Equal(t, original, result)
	})

	T.Run("status error without details returns original", func(t *testing.T) {
		t.Parallel()
		st := status.New(codes.NotFound, "not found")
		err := st.Err()
		result := DecodeErrorFromStatus(context.Background(), err)
		assert.Error(t, result)
	})

	T.Run("round-trips a platform sentinel error through encode/decode", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		original := platformerrors.ErrNilInputParameter

		// Encode using the interceptor helper
		detail := encodeErrorToDetails(ctx, original)
		require.NotNil(t, detail)

		// Build a status with details
		st := status.New(codes.InvalidArgument, original.Error())
		stWithDetails, err := st.WithDetails(detail)
		require.NoError(t, err)

		// Decode - the decoded error should contain the original message
		decoded := DecodeErrorFromStatus(ctx, stWithDetails.Err())
		require.Error(t, decoded)
		assert.Contains(t, decoded.Error(), "nil")
	})
}

func TestEncodeErrorToDetails(T *testing.T) {
	T.Parallel()

	T.Run("encodes a platform error", func(t *testing.T) {
		t.Parallel()
		detail := encodeErrorToDetails(context.Background(), platformerrors.ErrNilInputParameter)
		assert.NotNil(t, detail)
		assert.Equal(t, encodedErrorTypeURL, detail.TypeUrl)
	})

	T.Run("encodes a wrapped error", func(t *testing.T) {
		t.Parallel()
		wrapped := platformerrors.Wrap(platformerrors.ErrInvalidIDProvided, "context")
		detail := encodeErrorToDetails(context.Background(), wrapped)
		assert.NotNil(t, detail)
	})

	T.Run("encodes a simple error", func(t *testing.T) {
		t.Parallel()
		detail := encodeErrorToDetails(context.Background(), errors.New("simple"))
		// Even simple errors should encode (cockroachdb/errors handles them)
		assert.NotNil(t, detail)
	})
}

func TestUnaryErrorEncodingInterceptor(T *testing.T) {
	T.Parallel()

	T.Run("returns response when handler succeeds", func(t *testing.T) {
		t.Parallel()

		interceptor := UnaryErrorEncodingInterceptor()
		handler := func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		}

		resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, handler)
		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
	})

	T.Run("encodes platform error into status details", func(t *testing.T) {
		t.Parallel()

		interceptor := UnaryErrorEncodingInterceptor()
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, platformerrors.ErrNilInputParameter
		}

		resp, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, handler)
		assert.Nil(t, resp)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.NotEmpty(t, st.Details())
	})

	T.Run("preserves existing status code for known errors", func(t *testing.T) {
		t.Parallel()

		interceptor := UnaryErrorEncodingInterceptor()
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, sql.ErrNoRows
		}

		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	T.Run("handler returning status error preserves message", func(t *testing.T) {
		t.Parallel()

		interceptor := UnaryErrorEncodingInterceptor()
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, status.Error(codes.FailedPrecondition, "custom message")
		}

		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, "custom message", st.Message())
	})

	T.Run("unknown error uses codes.Unknown", func(t *testing.T) {
		t.Parallel()

		interceptor := UnaryErrorEncodingInterceptor()
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, errors.New("something unexpected")
		}

		_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unknown, st.Code())
	})
}

// mockServerStream implements grpc.ServerStream for testing.
type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SendMsg(any) error            { return nil }
func (m *mockServerStream) RecvMsg(any) error            { return nil }

func TestStreamErrorEncodingInterceptor(T *testing.T) {
	T.Parallel()

	T.Run("returns nil when handler succeeds", func(t *testing.T) {
		t.Parallel()

		interceptor := StreamErrorEncodingInterceptor()
		handler := func(srv any, stream grpc.ServerStream) error {
			return nil
		}

		ss := &mockServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, handler)
		assert.NoError(t, err)
	})

	T.Run("encodes platform error into status details", func(t *testing.T) {
		t.Parallel()

		interceptor := StreamErrorEncodingInterceptor()
		handler := func(srv any, stream grpc.ServerStream) error {
			return platformerrors.ErrInvalidIDProvided
		}

		ss := &mockServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.NotEmpty(t, st.Details())
	})

	T.Run("unknown error uses codes.Unknown", func(t *testing.T) {
		t.Parallel()

		interceptor := StreamErrorEncodingInterceptor()
		handler := func(srv any, stream grpc.ServerStream) error {
			return errors.New("stream failure")
		}

		ss := &mockServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unknown, st.Code())
	})

	T.Run("handler returning status error preserves message", func(t *testing.T) {
		t.Parallel()

		interceptor := StreamErrorEncodingInterceptor()
		handler := func(srv any, stream grpc.ServerStream) error {
			return status.Error(codes.Unauthenticated, "not authed")
		}

		ss := &mockServerStream{ctx: context.Background()}
		err := interceptor(nil, ss, &grpc.StreamServerInfo{}, handler)
		require.Error(t, err)

		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, "not authed", st.Message())
	})
}
