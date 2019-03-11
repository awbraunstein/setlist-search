package internal

import (
	"github.com/awbraunstein/setlist-search/index"
	"github.com/labstack/echo/v4"
)

const (
	// InjectorContextKey is the key used to lookup the Index from the echo.Context.
	InjectorContextKey = "index-injector-context-key"
)

// IndexInjector stores a pointer to an index and injects it into the context.
type IndexInjector struct {
	idx *index.Index
}

// NewInjector returns a new IndexInjector.
func NewInjector(location string) (*IndexInjector, error) {
	idx, err := index.Open(location)
	if err != nil {
		return nil, err
	}
	return &IndexInjector{idx: idx}, nil
}

// Middleware injects the index into the context.
func (s *IndexInjector) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(InjectorContextKey, s.idx)
		return next(c)
	}
}
