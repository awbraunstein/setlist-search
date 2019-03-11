package handlers

import (
	"net/http"

	echotrace "github.com/awbraunstein/echo-trace"
	"github.com/awbraunstein/setlist-search/internal"
	"github.com/labstack/echo"
	"golang.org/x/net/trace"
)

// SearchRequest is the json request to the api/search endpoint.
type SearchRequest struct {
	Query string `json:"query"`
}

func SearchAPI(c echo.Context) error {
	var req SearchRequest
	if err := internal.MergeJSONBody(c, &req); err != nil {
		tr := c.Get(echotrace.ContextKey).(trace.Trace)
		tr.LazyPrintf("Error parsing JSON: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if req.Query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing query param")
	}
	sr, err := searchIndex(c, req.Query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sr)
}
