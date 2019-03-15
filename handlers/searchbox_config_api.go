package handlers

import (
	"net/http"
	"sort"

	"github.com/awbraunstein/setlist-search/index"
	"github.com/awbraunstein/setlist-search/internal"
	"github.com/labstack/echo/v4"
)

type value struct {
	Data string `json:"data"`
	Text string `json:"text,omitempty"`
}

type valueKind struct {
	Name   string  `json:"name"`
	Color  string  `json:"color"`
	Values []value `json:"values"`
}

type searchboxConfig struct {
	ValueKinds []valueKind `json:"valueKinds"`
}

func SearchBoxConfigAPI(c echo.Context) error {
	idx := c.Get(internal.InjectorContextKey).(*index.Index)
	charsKind := valueKind{
		Name:  "special-characters",
		Color: "green",
		Values: []value{
			{Data: ")"},
			{Data: "("},
		},
	}

	keywordsKind := valueKind{
		Name:  "keywords",
		Color: "blue",
		Values: []value{
			{Data: "AND"},
			{Data: "OR"},
			{Data: "NOT"},
		},
	}

	songsKind := valueKind{
		Name:  "songs",
		Color: "red",
	}

	songSet := idx.Songs()
	var songLongNames []string
	for longName := range songSet {
		songLongNames = append(songLongNames, longName)
	}

	sort.Strings(songLongNames)
	for _, longName := range songLongNames {
		songsKind.Values = append(songsKind.Values, value{Text: longName, Data: songSet[longName]})
	}

	sbconf := searchboxConfig{
		ValueKinds: []valueKind{
			charsKind,
			keywordsKind,
			songsKind,
		},
	}

	resp := c.Response()
	header := resp.Header()
	// Allow clients to cache this since it changes infrequently.
	header.Add("Cache-Control", "private")
	// Allow it to be cached for up to 1 day.
	header.Add("Cache-Control", "max-age=86400")
	return c.JSON(http.StatusOK, &sbconf)
}
