package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

type homeTemplateData struct {
	SampleQueries []string
}

var homeData = &homeTemplateData{SampleQueries: sampleQueries}

func Home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl", homeData)
}
