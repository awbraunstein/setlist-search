package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

type homeTemplateData struct {
	SampleQueries []string
}

var homeData = &homeTemplateData{SampleQueries: []string{
	"farmhouse",
	"punch-you-in-the-eye AND fee AND the-sloth",
	"i-am-hydrogen AND (NOT mikes-song AND NOT weekapaug-groove)",
}}

func Home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl", homeData)
}
