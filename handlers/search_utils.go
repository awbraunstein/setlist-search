package handlers

import (
	"net/http"
	"sort"
	"time"

	echotrace "github.com/awbraunstein/echo-trace"
	"github.com/awbraunstein/setlist-search/index"
	"github.com/awbraunstein/setlist-search/internal"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/trace"
)

// SearchResults is the json payload for a search query.
type SearchResults struct {
	// Exported to the api.
	Count int      `json:"count"`
	Dates []string `json:"dates"`

	// Internal only.
	QueryTime time.Duration `json:"-"`
}

func searchIndex(c echo.Context, query string) (*SearchResults, error) {
	idx := c.Get(internal.InjectorContextKey).(*index.Index)
	start := time.Now()
	shows, err := idx.Query(query)
	if err != nil {
		tr := c.Get(echotrace.ContextKey).(trace.Trace)
		tr.LazyPrintf("Error executing query: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
	}
	elapsed := time.Since(start)
	tr := c.Get(echotrace.ContextKey).(trace.Trace)
	tr.LazyPrintf("Query %q completed in %v", query, elapsed)
	sr := &SearchResults{
		QueryTime: elapsed,
	}
	sr.Count = len(shows)
	for _, show := range shows {
		sr.Dates = append(sr.Dates, idx.ShowDate(show))
	}
	sort.Strings(sr.Dates)
	return sr, nil
}
