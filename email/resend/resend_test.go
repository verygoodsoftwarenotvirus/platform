package resend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cbnoop "github.com/verygoodsoftwarenotvirus/platform/v5/circuitbreaking/noop"
	"github.com/verygoodsoftwarenotvirus/platform/v5/email"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v5/observability/tracing"

	"github.com/stretchr/testify/require"
)

type sendEmailResponse struct {
	Id string `json:"id"`
}

func TestNewResendEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{APIToken: t.Name()}

		client, err := NewResendEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	T.Run("with missing config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		client, err := NewResendEmailer(nil, logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing config API token", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{}

		client, err := NewResendEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, cbnoop.NewCircuitBreaker(), nil)
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing HTTP client", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{APIToken: t.Name()}

		client, err := NewResendEmailer(config, logger, tracing.NewNoopTracerProvider(), nil, cbnoop.NewCircuitBreaker(), nil)
		require.Nil(t, client)
		require.Error(t, err)
	})
}

func TestResendEmailer_SendEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			json.NewEncoder(res).Encode(sendEmailResponse{Id: t.Name()})
		}))

		cfg := &Config{APIToken: t.Name()}

		c, err := NewResendEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), cbnoop.NewCircuitBreaker(), nil)
		require.NotNil(t, c)
		require.NoError(t, err)

		baseURL, err := url.Parse(ts.URL + "/")
		require.NoError(t, err)
		c.client.BaseURL = baseURL

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

		cfg := &Config{APIToken: t.Name()}

		c, err := NewResendEmailer(cfg, logger, tracing.NewNoopTracerProvider(), client, cbnoop.NewCircuitBreaker(), nil)
		require.NotNil(t, c)
		require.NoError(t, err)

		baseURL, err := url.Parse(ts.URL + "/")
		require.NoError(t, err)
		c.client.BaseURL = baseURL

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

		cfg := &Config{APIToken: t.Name()}

		c, err := NewResendEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), cbnoop.NewCircuitBreaker(), nil)
		require.NotNil(t, c)
		require.NoError(t, err)

		baseURL, err := url.Parse(ts.URL + "/")
		require.NoError(t, err)
		c.client.BaseURL = baseURL

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
