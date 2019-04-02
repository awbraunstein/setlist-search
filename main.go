package main

import (
	"context"
	"flag"
	"io"
	"os"
	"path/filepath"
	"text/template"

	echotrace "github.com/awbraunstein/echo-trace"
	"github.com/awbraunstein/setlist-search/handlers"
	"github.com/awbraunstein/setlist-search/internal"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	httpAddr    = flag.String("http", ":8080", "Listen address")
	remoteIndex = flag.Bool("remote_index", true, "Whether the index should be fetched from the remote source")
)

func getIndexLocation() string {
	if indexLocation := os.Getenv("SETSEARCHERINDEX"); indexLocation != "" {
		return indexLocation
	}
	return filepath.Clean(os.Getenv("HOME") + "/.setsearcherindex")
}

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates[name].ExecuteTemplate(w, "base", data)
}

func parseTemplates() *Template {
	libs, err := filepath.Glob("templates/library/*.tmpl")
	if err != nil {
		panic(err.Error())
	}
	tmpls, err := filepath.Glob("templates/*.tmpl")
	if err != nil {
		panic(err.Error())
	}
	t := &Template{templates: make(map[string]*template.Template)}
	for _, fname := range tmpls {
		command := filepath.Base(fname)
		t.templates[command] = template.Must(template.ParseFiles(append([]string{fname}, libs...)...))
	}
	return t
}

// This is the entrypoint into the setlist server.
func main() {
	flag.Parse()
	e := echo.New()
	e.Renderer = parseTemplates()

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(echotrace.Middleware)
	var injector *internal.IndexInjector
	var err error
	if *remoteIndex {
		injector, err = internal.NewCloudInjector(context.Background(), "setlist-searcher-index", "index.txt")
	} else {
		injector, err = internal.NewInjector(getIndexLocation())
	}
	if err != nil {
		e.Logger.Fatal(err)
	}
	e.Use(injector.Middleware)

	e.GET("/", handlers.Home)
	e.GET("/search", handlers.Search)

	// Accept /api/search on GET and POST.
	e.GET("/api/search", handlers.SearchAPI)
	e.POST("/api/search", handlers.SearchAPI)

	// Accept /api/search on GET.
	e.GET("/api/searchboxconfig", handlers.SearchBoxConfigAPI)

	e.GET("/debug/requests", echotrace.Handler)

	e.Static("/static", "assets")
	e.Logger.Fatal(e.Start(*httpAddr))
}
