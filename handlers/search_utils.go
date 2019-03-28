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

type ShowInfo struct {
	Date string `json:"date"`
	Url  string `json:"url"`
}

type byDate []ShowInfo

func (si byDate) Len() int {
	return len(si)
}
func (si byDate) Swap(i, j int) {
	si[i], si[j] = si[j], si[i]
}
func (si byDate) Less(i, j int) bool {
	return si[i].Date < si[j].Date
}

// SearchResults is the json payload for a search query.
type SearchResults struct {
	// Exported to the api.
	Count int        `json:"count"`
	Shows []ShowInfo `json:"shows"`

	// Internal only.
	QueryTime time.Duration `json:"-"`
}

func searchIndex(c echo.Context, query string) (*SearchResults, error) {
	idx := c.Get(internal.InjectorContextKey).(*index.Index)
	start := time.Now()
	shows, err := idx.Query(c.Request().Context(), query)
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
		sr.Shows = append(sr.Shows, ShowInfo{
			Date: idx.ShowDate(show),
			Url:  idx.ShowUrl(show),
		})
	}
	sort.Sort(byDate(sr.Shows))
	return sr, nil
}
