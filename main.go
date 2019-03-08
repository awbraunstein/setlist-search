package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	echotrace "github.com/awbraunstein/echo-trace"
	"github.com/awbraunstein/setlist-search/index"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/trace"
)

var (
	httpAddr = flag.String("http", ":8080", "Listen address")
)

type server struct {
	*echo.Echo

	indx *index.Index
}

func getIndexLocation() string {
	if indexLocation := os.Getenv("SETSEARCHERINDEX"); indexLocation != "" {
		return indexLocation
	}
	return filepath.Clean(os.Getenv("HOME") + "/.setsearcherindex")
}

func newServer() (*server, error) {
	indx, err := index.Open(getIndexLocation())
	if err != nil {
		return nil, err
	}

	return &server{
		Echo: echo.New(),
		indx: indx,
	}, nil
}

// SearchResult is the json payload for a search query.
type SearchResult struct {
	Count int      `json:"count"`
	Dates []string `json:"dates"`
}

// This is the entrypoint into the setlist server.
func main() {

	flag.Parse()
	e, err := newServer()
	if err != nil {
		log.Fatalf("Unable to create server: %v", err)
	}
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(echotrace.Middleware)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/api/search", func(c echo.Context) error {
		query := c.QueryParam("query")
		if query == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing query param")
		}
		start := time.Now()
		shows, err := e.indx.Query(query)
		if err != nil {
			tr := c.Get(echotrace.ContextKey).(trace.Trace)
			tr.LazyPrintf("Error executing query: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Error")
		}
		elapsed := time.Since(start)
		tr := c.Get(echotrace.ContextKey).(trace.Trace)
		tr.LazyPrintf("Query %q completed in %v", query, elapsed)
		sr := &SearchResult{}
		sr.Count = len(shows)
		for _, show := range shows {
			sr.Dates = append(sr.Dates, e.indx.ShowDate(show))
		}
		sort.Strings(sr.Dates)
		return c.JSON(http.StatusOK, sr)
	})

	e.GET("/debug/requests", echotrace.Handler)

	e.Logger.Fatal(e.Start(*httpAddr))
}
