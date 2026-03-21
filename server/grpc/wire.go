package grpc

import (
	"github.com/google/wire"
)

var (
	ProvidersGRPC = wire.NewSet(
		NewGRPCServer,
	)
)
