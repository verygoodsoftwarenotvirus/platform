package filtering

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v3/observability/logging"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/v3/search/text"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryFilter_AttachToLogger(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		qf := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
			CreatedAfter:    new(time.Now().Truncate(time.Second)),
			CreatedBefore:   new(time.Now().Truncate(time.Second)),
			UpdatedAfter:    new(time.Now().Truncate(time.Second)),
			UpdatedBefore:   new(time.Now().Truncate(time.Second)),
			SortBy:          SortDescending,
			IncludeArchived: new(true),
		}

		assert.NotNil(t, qf.AttachToLogger(logger))
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()

		assert.NotNil(t, (*QueryFilter)(nil).AttachToLogger(logger))
	})
}

func TestQueryFilter_FromParams(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tt, err := time.Parse(time.RFC3339Nano, time.Now().UTC().Truncate(time.Second).Format(time.RFC3339Nano))
		require.NoError(t, err)

		actual := &QueryFilter{}
		expected := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
			CreatedAfter:    new(tt),
			CreatedBefore:   new(tt),
			UpdatedAfter:    new(tt),
			UpdatedBefore:   new(tt),
			SortBy:          SortDescending,
			IncludeArchived: new(true),
		}

		exampleInput := url.Values{
			textsearch.QueryKeySearch: []string{t.Name()},
			QueryKeyCursor:            []string{*expected.Cursor},
			QueryKeyLimit:             []string{strconv.Itoa(int(*expected.MaxResponseSize))},
			QueryKeyCreatedBefore:     []string{expected.CreatedAfter.Format(time.RFC3339Nano)},
			QueryKeyCreatedAfter:      []string{expected.CreatedBefore.Format(time.RFC3339Nano)},
			QueryKeyUpdatedBefore:     []string{expected.UpdatedAfter.Format(time.RFC3339Nano)},
			QueryKeyUpdatedAfter:      []string{expected.UpdatedBefore.Format(time.RFC3339Nano)},
			QueryKeySortBy:            []string{*expected.SortBy},
			QueryKeyIncludeArchived:   []string{strconv.FormatBool(true)},
		}

		actual.FromParams(exampleInput)

		assert.Equal(t, expected, actual)

		exampleInput[QueryKeySortBy] = []string{*SortAscending}

		actual.FromParams(exampleInput)
		assert.Equal(t, SortAscending, actual.SortBy)
	})
}

func TestQueryFilter_SetCursor(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := t.Name()
		qf := &QueryFilter{}
		qf.SetCursor(&expected)

		assert.Equal(t, expected, *qf.Cursor)
	})
}

func TestQueryFilter_ToValues(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tt, err := time.Parse(time.RFC3339Nano, time.Now().UTC().Truncate(time.Second).Format(time.RFC3339Nano))
		require.NoError(t, err)

		qf := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
			CreatedAfter:    new(tt),
			CreatedBefore:   new(tt),
			UpdatedAfter:    new(tt),
			UpdatedBefore:   new(tt),
			SortBy:          SortDescending,
			IncludeArchived: new(true),
		}

		expected := url.Values{
			QueryKeyCursor:          []string{*qf.Cursor},
			QueryKeyLimit:           []string{strconv.Itoa(int(*qf.MaxResponseSize))},
			QueryKeyCreatedBefore:   []string{qf.CreatedAfter.Format(time.RFC3339Nano)},
			QueryKeyCreatedAfter:    []string{qf.CreatedBefore.Format(time.RFC3339Nano)},
			QueryKeyUpdatedBefore:   []string{qf.UpdatedAfter.Format(time.RFC3339Nano)},
			QueryKeyUpdatedAfter:    []string{qf.UpdatedBefore.Format(time.RFC3339Nano)},
			QueryKeyIncludeArchived: []string{strconv.FormatBool(*qf.IncludeArchived)},
			QueryKeySortBy:          []string{*qf.SortBy},
		}

		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()
		qf := (*QueryFilter)(nil)
		expected := DefaultQueryFilter().ToValues()
		actual := qf.ToValues()
		assert.Equal(t, expected, actual)
	})
}

func TestExtractQueryFilter(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		tt, err := time.Parse(time.RFC3339Nano, time.Now().UTC().Truncate(time.Second).Format(time.RFC3339Nano))
		require.NoError(t, err)

		expected := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
			CreatedAfter:    new(tt),
			CreatedBefore:   new(tt),
			UpdatedAfter:    new(tt),
			UpdatedBefore:   new(tt),
			SortBy:          SortDescending,
		}
		exampleInput := url.Values{
			textsearch.QueryKeySearch: []string{t.Name()},
			QueryKeyCursor:            []string{*expected.Cursor},
			QueryKeyLimit:             []string{strconv.Itoa(int(*expected.MaxResponseSize))},
			QueryKeyCreatedBefore:     []string{expected.CreatedAfter.Format(time.RFC3339Nano)},
			QueryKeyCreatedAfter:      []string{expected.CreatedBefore.Format(time.RFC3339Nano)},
			QueryKeyUpdatedBefore:     []string{expected.UpdatedAfter.Format(time.RFC3339Nano)},
			QueryKeyUpdatedAfter:      []string{expected.UpdatedBefore.Format(time.RFC3339Nano)},
			QueryKeySortBy:            []string{*expected.SortBy},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://verygoodsoftwarenotvirus.ru", http.NoBody)
		assert.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = exampleInput.Encode()
		actual := ExtractQueryFilterFromRequest(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("with missing values", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(DefaultQueryFilterLimit)),
			SortBy:          SortAscending,
		}
		exampleInput := url.Values{
			QueryKeyCursor: []string{*expected.Cursor},
			QueryKeyLimit:  []string{"0"},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://verygoodsoftwarenotvirus.ru", http.NoBody)
		assert.NoError(t, err)
		require.NotNil(t, req)

		req.URL.RawQuery = exampleInput.Encode()
		actual := ExtractQueryFilterFromRequest(req)
		assert.Equal(t, expected, actual)
	})
}

func TestQueryFilter_ToPagination(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		qf := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
		}

		expected := Pagination{
			Cursor:          *qf.Cursor,
			MaxResponseSize: *qf.MaxResponseSize,
		}

		actual := qf.ToPagination()
		assert.Equal(t, expected, actual)
	})

	T.Run("with nil value", func(t *testing.T) {
		t.Parallel()

		qf := (*QueryFilter)(nil)

		actual := qf.ToPagination()
		assert.NotNil(t, actual)
	})
}

func TestNewQueryFilteredResult(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		qf := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
		}

		data := []*string{new("a"), new("b")}
		filteredCount := uint64(len(data))
		totalCount := uint64(len(data))
		idExtractor := func(s *string) string { return *s }

		expected := &QueryFilteredResult[string]{
			Data: data,
			Pagination: Pagination{
				Cursor:             *data[1],
				PreviousCursor:     *qf.Cursor,
				MaxResponseSize:    *qf.MaxResponseSize,
				FilteredCount:      filteredCount,
				TotalCount:         totalCount,
				AppliedQueryFilter: qf,
			},
		}

		actual := NewQueryFilteredResult(data, filteredCount, totalCount, idExtractor, qf)
		assert.Equal(t, expected, actual)
	})

	T.Run("with empty data", func(t *testing.T) {
		t.Parallel()

		qf := &QueryFilter{
			Cursor:          new(t.Name()),
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
		}

		data := []*string{}
		filteredCount := uint64(0)
		totalCount := uint64(0)
		idExtractor := func(s *string) string { return *s }

		expected := &QueryFilteredResult[string]{
			Data: data,
			Pagination: Pagination{
				Cursor:             "",
				PreviousCursor:     *qf.Cursor,
				MaxResponseSize:    *qf.MaxResponseSize,
				FilteredCount:      filteredCount,
				TotalCount:         totalCount,
				AppliedQueryFilter: qf,
			},
		}

		actual := NewQueryFilteredResult(data, filteredCount, totalCount, idExtractor, qf)
		assert.Equal(t, expected, actual)
	})

	T.Run("with no cursor", func(t *testing.T) {
		t.Parallel()

		qf := &QueryFilter{
			MaxResponseSize: new(uint8(MaxQueryFilterLimit)),
		}

		data := []*string{new("a"), new("b")}
		filteredCount := uint64(len(data))
		totalCount := uint64(len(data))
		idExtractor := func(s *string) string { return *s }

		expected := &QueryFilteredResult[string]{
			Data: data,
			Pagination: Pagination{
				Cursor:             *data[1],
				PreviousCursor:     "",
				MaxResponseSize:    *qf.MaxResponseSize,
				FilteredCount:      filteredCount,
				TotalCount:         totalCount,
				AppliedQueryFilter: qf,
			},
		}

		actual := NewQueryFilteredResult(data, filteredCount, totalCount, idExtractor, qf)
		assert.Equal(t, expected, actual)
	})
}
