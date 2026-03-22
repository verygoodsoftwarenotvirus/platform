package mockrouting

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v2/observability/logging"

	"github.com/stretchr/testify/mock"
)

// NewRouteParamManager returns a new RouteParamManager.
func NewRouteParamManager() *RouteParamManager {
	return &RouteParamManager{}
}

// RouteParamManager is a mock routing.RouteParamManager.
type RouteParamManager struct {
	mock.Mock
}

// UserIDFetcherFromSessionContextData satisfies our interface contract.
func (m *RouteParamManager) UserIDFetcherFromSessionContextData(req *http.Request) uint64 {
	return m.Called(req).Get(0).(uint64)
}

// BuildRouteParamIDFetcher satisfies our interface contract.
func (m *RouteParamManager) BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(*http.Request) uint64 {
	return m.Called(logger, key, logDescription).Get(0).(func(*http.Request) uint64)
}

// BuildRouteParamStringIDFetcher satisfies our interface contract.
func (m *RouteParamManager) BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string {
	return m.Called(key).Get(0).(func(*http.Request) string)
}
