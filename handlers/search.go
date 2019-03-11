package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

type searchTemplateData struct {
	Query   string
	Results *SearchResults
}

func Search(c echo.Context) error {
	query := c.QueryParam("query")
	if query == "" {
		c.Redirect(http.StatusMovedPermanently, "/")
	}
	sr, _ := searchIndex(c, query)
	data := &searchTemplateData{
		Query:   query,
		Results: sr,
	}
	return c.Render(http.StatusOK, "search.tmpl", data)
}
