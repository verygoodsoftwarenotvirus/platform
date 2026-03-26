package mobilenotifications

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestMultiPlatformPushSender_SendPush(T *testing.T) {
	T.Parallel()

	ctx := T.Context()
	logger := logging.NewNoopLogger()
	tracer := tracing.NewNoopTracerProvider()

	T.Run("ios returns ErrPlatformNotSupported when apnsSender nil", func(t *testing.T) {
		t.Parallel()

		sender := NewMultiPlatformPushSender(nil, nil, logger, tracer)
		err := sender.SendPush(ctx, "ios", "token", PushMessage{Title: "title", Body: "body"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPlatformNotSupported)
	})

	T.Run("android returns ErrPlatformNotSupported when fcmSender nil", func(t *testing.T) {
		t.Parallel()

		sender := NewMultiPlatformPushSender(nil, nil, logger, tracer)
		err := sender.SendPush(ctx, "android", "token", PushMessage{Title: "title", Body: "body"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPlatformNotSupported)
	})

	T.Run("unknown platform returns error", func(t *testing.T) {
		t.Parallel()

		sender := NewMultiPlatformPushSender(nil, nil, logger, tracer)
		err := sender.SendPush(ctx, "unknown", "token", PushMessage{Title: "title", Body: "body"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown platform")
	})
}
