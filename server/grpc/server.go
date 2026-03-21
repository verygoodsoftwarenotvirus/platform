package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/verygoodsoftwarenotvirus/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "grpc_server"
)

type (
	Config struct {
		HTTPSCertificateFile  string `env:"TLS_CERTIFICATE_FILEPATH"     json:"tlsCertificate,omitempty"`
		TLSCertificateKeyFile string `env:"TLS_CERTIFICATE_KEY_FILEPATH" json:"tlsCertificateKey,omitempty"`
		Port                  uint16 `env:"PORT"                         json:"port"`
	}

	Server struct {
		logger         logging.Logger
		config         *Config
		grpcServer     *grpc.Server
		tracerProvider tracing.TracerProvider
	}

	// RegistrationFunc is i.e. protobuf.RegisterSomeExampleServiceServer(grpcServer, &exampleServiceServerImpl{}).
	RegistrationFunc func(*grpc.Server)
)

func NewGRPCServer(
	cfg *Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	unaryServerInterceptors []grpc.UnaryServerInterceptor,
	streamServerInterceptors []grpc.StreamServerInterceptor,
	registrationFunctions ...RegistrationFunc,
) (*Server, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("grpc server")
	}

	tp := tracing.EnsureTracerProvider(tracerProvider)
	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(tp))),
		grpc.ChainUnaryInterceptor(append([]grpc.UnaryServerInterceptor{LoggingInterceptor(logger)}, unaryServerInterceptors...)...),
		grpc.ChainStreamInterceptor(streamServerInterceptors...),
	}

	if cfg.TLSCertificateKeyFile != "" && cfg.HTTPSCertificateFile != "" {
		serverCert, err := tls.LoadX509KeyPair(cfg.HTTPSCertificateFile, cfg.TLSCertificateKeyFile)
		if err != nil {
			return nil, err
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.NoClientCert,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(config)))
	}

	grpcServer := grpc.NewServer(opts...)
	for _, rf := range registrationFunctions {
		rf(grpcServer)
	}

	reflection.Register(grpcServer)

	return &Server{
		logger:         logging.EnsureLogger(logger).WithName(serviceName),
		config:         cfg,
		grpcServer:     grpcServer,
		tracerProvider: tp,
	}, nil
}

// Shutdown shuts down the server. Call with a context that has sufficient timeout
// to allow in-flight spans to be flushed to the collector.
func (s *Server) Shutdown(ctx context.Context) {
	if err := s.tracerProvider.ForceFlush(ctx); err != nil {
		s.logger.Error("flushing traces", err)
	}
	s.grpcServer.Stop()
}

// Serve serves GRPC traffic.
func (s *Server) Serve() {
	var lc net.ListenConfig
	lis, err := lc.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s.logger.WithValue("port", s.config.Port).Info("Listening for GRPC requests")
	if err = s.grpcServer.Serve(lis); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			// NOTE: there is a chance that next line won't have tim  e to run,
			// as main() doesn't wait for this goroutine to stop.
			os.Exit(0)
		}
	}
}

func LoggingInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	l := logging.EnsureLogger(logger)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var ev uint8
		result, err := handler(ctx, req)
		if err != nil {
			ev = 1
		}
		l.WithValue("rpc.method", info.FullMethod).WithValue("err", ev).Info("rpc invoked")

		return result, err
	}
}
