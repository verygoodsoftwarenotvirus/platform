package capitalism

import (
	"net/http"

	"github.com/dinnerdonebetter/backend/internal/platform/observability/tracing"

	servertiming "github.com/mitchellh/go-server-timing"
)

// IncomingWebhookHandler is our webhook handling route.
func (s *service) IncomingWebhookHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	timing := servertiming.FromContext(ctx)
	logger := s.logger.WithRequest(req).WithSpan(span)
	tracing.AttachRequestToSpan(span, req)

	webhookHandleTimer := timing.NewMetric("handler").WithDesc("handle webhook event").Start()
	statusToWrite := http.StatusOK
	if err := s.paymentManager.HandleEventWebhook(req); err != nil {
		logger.Error("handling subscription event webhook", err)
		statusToWrite = http.StatusInternalServerError
	}
	webhookHandleTimer.Stop()

	res.WriteHeader(statusToWrite)
}
