package sendgrid

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v3/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/platform/v3/email"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/tracing"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stretchr/testify/require"
)

func TestNewSendGridEmailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		client, err := NewSendGridEmailer(&Config{APIToken: t.Name()}, logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, client)
		require.NoError(t, err)
	})
}

func TestSendGridEmailer_SendEmail(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusAccepted)
		}))

		c, err := NewSendGridEmailer(&Config{APIToken: t.Name()}, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.client.BaseURL = ts.URL

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
		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Hour)
		}))
		client := ts.Client()
		client.Timeout = time.Millisecond

		c, err := NewSendGridEmailer(&Config{APIToken: t.Name()}, logger, tracing.NewNoopTracerProvider(), client, circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.client.BaseURL = ts.URL

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
		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusInternalServerError)
		}))

		c, err := NewSendGridEmailer(&Config{APIToken: t.Name()}, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.client.BaseURL = ts.URL

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

func TestSendGridEmailer_sendDynamicTemplateEmail(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		logger := logging.NewNoopLogger()

		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusAccepted)
		}))

		c, err := NewSendGridEmailer(&Config{APIToken: t.Name()}, logger, tracing.NewNoopTracerProvider(), ts.Client(), circuitbreaking.NewNoopCircuitBreaker())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.client.BaseURL = ts.URL

		ctx := t.Context()
		to := mail.NewEmail("sender", "sender@fake.com")
		from := mail.NewEmail("sender", "sender@fake.com")

		request := sendgrid.GetRequest(c.config.APIToken, "/v3/mail/send", ts.URL)

		require.NoError(t, c.sendDynamicTemplateEmail(ctx, to, from, t.Name(), map[string]any{"things": "stuff"}, request))
	})
}
