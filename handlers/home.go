package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type sampleQuery struct {
	Query      string
	HumanValue string
}

type homeTemplateData struct {
	SampleQueries []sampleQuery
}

var homeData = &homeTemplateData{SampleQueries: []sampleQuery{
	{"farmhouse", "Farmhouse"},
	{"punch-you-in-the-eye AND fee AND the-sloth", "Punch You in the Eye AND The Sloth"},
	{"i-am-hydrogen AND (NOT mikes-song AND NOT weekapaug-groove)", "I am Hydrogen AND (NOT Mike's Song AND NOT Weekapaug Groove)"},
}}

func Home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl", homeData)
}
