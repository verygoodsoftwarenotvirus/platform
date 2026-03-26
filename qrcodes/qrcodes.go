package qrcodes

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

const (
	o11yName          = "qr_code_builder"
	base64ImagePrefix = "data:image/jpeg;base64,"
)

type (
	// Builder generates QR codes for TOTP two-factor authentication.
	Builder interface {
		BuildQRCode(ctx context.Context, username, twoFactorSecret string) string
	}

	// Issuer identifies the service that issued the TOTP secret.
	Issuer string

	builder struct {
		tracer     tracing.Tracer
		logger     logging.Logger
		totpIssuer Issuer
	}
)

// NewBuilder returns a new QR code Builder.
func NewBuilder(tracerProvider tracing.TracerProvider, logger logging.Logger) Builder {
	return &builder{
		tracer:     tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		logger:     logging.EnsureLogger(logger).WithName(o11yName),
		totpIssuer: "UNKNOWN_REPLACEME",
	}
}

// BuildQRCode builds a QR code for a given username and secret.
func (s *builder) BuildQRCode(ctx context.Context, username, twoFactorSecret string) string {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.UsernameKey, username)

	// "otpauth://totp/{{ .Issuer }}:{{ .EnsureUsername }}?secret={{ .Secret }}&issuer={{ .Issuer }}",
	otpString := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		s.totpIssuer,
		username,
		twoFactorSecret,
		s.totpIssuer,
	)

	// encode two factor secret as authenticator-friendly QR code
	qrCode, err := qr.Encode(otpString, qr.L, qr.Auto)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding OTP string")
		return ""
	}

	// scale the QR code so that it's not a PNG for ants.
	qrCode, err = barcode.Scale(qrCode, 256, 256)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "scaling QR code")
		return ""
	}

	// encode the QR code to PNG.
	var b bytes.Buffer
	if err = png.Encode(&b, qrCode); err != nil {
		observability.AcknowledgeError(err, logger, span, "encoding QR code to PNG")
		return ""
	}

	// base64 encode the image for easy HTML use.
	return fmt.Sprintf("%s%s", base64ImagePrefix, base64.StdEncoding.EncodeToString(b.Bytes()))
}
