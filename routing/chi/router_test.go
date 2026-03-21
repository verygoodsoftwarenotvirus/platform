package chi

import (
	"context"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/routing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRouterForTest() routing.Router {
	return NewRouter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{})
}

func TestNewRouter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, NewRouter(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{}))
	})
}

func Test_buildChiMux(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, buildChiMux(logging.NewNoopLogger(), tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer(t.Name())), metrics.NewNoopMetricsProvider(), &Config{}))
	})
}

func Test_convertMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, convertMiddleware(func(http.Handler) http.Handler { return nil }))
	})
}

func Test_router_AddRoute(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		methods := []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}

		for _, method := range methods {
			assert.NoError(t, r.AddRoute(method, "/path", nil))
		}
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		assert.Error(t, r.AddRoute("blah", "/path", nil))
	})
}

func Test_router_Connect(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Connect("/test", nil)
	})
}

func Test_router_Delete(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Delete("/test", nil)
	})
}

func Test_router_Get(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Get("/test", nil)
	})
}

func Test_router_Handle(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Handle("/test", nil)
	})
}

func Test_router_HandleFunc(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.HandleFunc("/test", nil)
	})
}

func Test_router_Handler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		assert.NotNil(t, r.Handler())
	})
}

func Test_router_Head(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Head("/test", nil)
	})
}

func Test_router_LogRoutes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		assert.NoError(t, r.AddRoute(http.MethodGet, "/path", nil))

		r.Routes()
	})
}

func Test_router_Options(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Options("/test", nil)
	})
}

func Test_router_Patch(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Patch("/test", nil)
	})
}

func Test_router_Post(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Post("/test", nil)
	})
}

func Test_router_Put(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Put("/thing", nil)
	})
}

func Test_router_Route(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		assert.NotNil(t, r.Route("/test", func(routing.Router) {}))
	})
}

func Test_router_Trace(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		r.Trace("/test", nil)
	})
}

func Test_router_WithMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouterForTest()

		assert.NotNil(t, r.WithMiddleware())
	})
}

func Test_router_clone(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouter(nil, nil, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{})

		assert.NotNil(t, r.clone())
	})
}

func Test_router_BuildRouteParamIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouter(nil, nil, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{})
		l := logging.NewNoopLogger()
		ctx := t.Context()
		exampleKey := "blah"

		rf := r.BuildRouteParamIDFetcher(l, exampleKey, "desc")
		assert.NotNil(t, rf)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/blah", http.NoBody)
		assert.NoError(t, err)
		require.NotNil(t, req)

		expected := uint64(123456)

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
			URLParams: chi.RouteParams{
				Keys:   []string{exampleKey},
				Values: []string{"123456"},
			},
		}))

		actual := rf(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("without appropriate value attached to context", func(t *testing.T) {
		t.Parallel()

		r := buildRouter(nil, nil, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{})
		l := logging.NewNoopLogger()
		ctx := t.Context()
		exampleKey := "blah"

		rf := r.BuildRouteParamIDFetcher(l, exampleKey, "desc")
		assert.NotNil(t, rf)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/blah", http.NoBody)
		assert.NoError(t, err)
		require.NotNil(t, req)

		actual := rf(req)
		assert.Zero(t, actual)
	})
}

func Test_router_BuildRouteParamStringIDFetcher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := buildRouter(nil, nil, tracing.NewNoopTracerProvider(), metrics.NewNoopMetricsProvider(), &Config{})
		ctx := t.Context()
		exampleKey := "blah"

		rf := r.BuildRouteParamStringIDFetcher(exampleKey)
		assert.NotNil(t, rf)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/blah", http.NoBody)
		assert.NoError(t, err)
		require.NotNil(t, req)

		expected := "fake_user_id"

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
			URLParams: chi.RouteParams{
				Keys:   []string{exampleKey},
				Values: []string{expected},
			},
		}))

		actual := rf(req)
		assert.Equal(t, expected, actual)
	})
}
