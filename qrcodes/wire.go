package qrcodes

import "github.com/google/wire"

var (
	QRCodeProviders = wire.NewSet(
		NewBuilder,
	)
)
