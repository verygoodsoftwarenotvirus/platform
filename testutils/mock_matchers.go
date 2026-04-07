package testutils

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/platform/v5/database/filtering"

	"github.com/stretchr/testify/mock"
)

// MatchType is a matcher for use with testify/mock.
func MatchType[T any]() any {
	return mock.MatchedBy(func(T) bool {
		return true
	})
}

var (
	ContextMatcher            = MatchType[context.Context]()
	QueryFilterMatcher        = MatchType[*filtering.QueryFilter]()
	HTTPRequestMatcher        = MatchType[*http.Request]()
	HTTPResponseWriterMatcher = MatchType[http.ResponseWriter]()
)
