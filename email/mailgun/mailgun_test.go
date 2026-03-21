package mailgun

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/email"
	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"

	"github.com/stretchr/testify/require"
)

const (
	exampleDomain = "dinnerdonebetter.dev"
)

type sendMessageResponse struct {
	Message string `json:"message"`
	Id      string `json:"id"`
}

func TestNewMailgunEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{Domain: exampleDomain, PrivateAPIKey: t.Name()}

		client, err := NewMailgunEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	T.Run("with missing config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		client, err := NewMailgunEmailer(nil, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing config domain", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{PrivateAPIKey: t.Name()}

		client, err := NewMailgunEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing config private key", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{Domain: exampleDomain}

		client, err := NewMailgunEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing HTTP client", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{Domain: exampleDomain, PrivateAPIKey: t.Name()}

		client, err := NewMailgunEmailer(config, logger, tracing.NewNoopTracerProvider(), nil, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})
}

func TestMailgunEmailer_SendEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			json.NewEncoder(res).Encode(sendMessageResponse{
				Message: "Queued. Thank you.",
				Id:      t.Name(),
			})
		}))

		cfg := &Config{Domain: exampleDomain, PrivateAPIKey: t.Name()}

		c, err := NewMailgunEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.client.SetAPIBase(ts.URL + "/v4")

		ctx := t.Context()
		details := &email.OutboundEmailMessage{
			ToAddress:   t.Name(),
			ToName:      t.Name(),
			FromAddress: t.Name(),
			FromName:    t.Name(),
			Subject:     t.Name(),
			HTMLContent: t.Name(),
		}

		require.NoError(t, c.SendEmail(ctx, details))
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Hour)
		}))

		client := ts.Client()
		client.Timeout = time.Millisecond

		cfg := &Config{Domain: exampleDomain, PrivateAPIKey: t.Name()}

		c, err := NewMailgunEmailer(cfg, logger, tracing.NewNoopTracerProvider(), client, circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)
		ctx := t.Context()
		details := &email.OutboundEmailMessage{
			ToAddress:   t.Name(),
			ToName:      t.Name(),
			FromAddress: t.Name(),
			FromName:    t.Name(),
			Subject:     t.Name(),
			HTMLContent: t.Name(),
		}

		err = c.SendEmail(ctx, details)
		require.Error(t, err)
	})

	T.Run("with invalid response code", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusInternalServerError)
		}))

		cfg := &Config{Domain: exampleDomain, PrivateAPIKey: t.Name()}

		c, err := NewMailgunEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		ctx := t.Context()
		details := &email.OutboundEmailMessage{
			ToAddress:   t.Name(),
			ToName:      t.Name(),
			FromAddress: t.Name(),
			FromName:    t.Name(),
			Subject:     t.Name(),
			HTMLContent: t.Name(),
		}

		err = c.SendEmail(ctx, details)
		require.Error(t, err)
	})
}
