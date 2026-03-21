package postmark

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

type emailResponse struct {
	Message     string `json:"Message"`
	MessageID   string `json:"MessageID"`
	SubmittedAt string `json:"SubmittedAt"`
	To          string `json:"To"`
	ErrorCode   int64  `json:"ErrorCode"`
}

func TestNewPostmarkEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{ServerToken: t.Name()}

		client, err := NewPostmarkEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, client)
		require.NoError(t, err)
	})

	T.Run("with missing config", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		client, err := NewPostmarkEmailer(nil, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing server token", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{}

		client, err := NewPostmarkEmailer(config, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})

	T.Run("with missing HTTP client", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		config := &Config{ServerToken: t.Name()}

		client, err := NewPostmarkEmailer(config, logger, tracing.NewNoopTracerProvider(), nil, circuitbreaking.NewNoopCircuitBreaker())
		require.Nil(t, client)
		require.Error(t, err)
	})
}

func TestPostmarkEmailer_SendEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			json.NewEncoder(res).Encode(emailResponse{
				ErrorCode:   0,
				Message:     "OK",
				MessageID:   t.Name(),
				SubmittedAt: "2010-11-26T12:01:05-05:00",
				To:          t.Name(),
			})
		}))

		cfg := &Config{ServerToken: t.Name(), BaseURL: ts.URL}

		c, err := NewPostmarkEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
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

		cfg := &Config{ServerToken: t.Name(), BaseURL: ts.URL}

		c, err := NewPostmarkEmailer(cfg, logger, tracing.NewNoopTracerProvider(), client, circuitbreaking.NewNoopCircuitBreaker())
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

		cfg := &Config{ServerToken: t.Name(), BaseURL: ts.URL}

		c, err := NewPostmarkEmailer(cfg, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
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
