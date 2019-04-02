package internal

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"github.com/awbraunstein/setlist-search/index"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	// InjectorContextKey is the key used to lookup the Index from the echo.Context.
	InjectorContextKey = "index-injector-context-key"
)

// IndexInjector stores a pointer to an index and injects it into the context.
type IndexInjector struct {
	idx *index.Index
}

func NewCloudInjector(ctx context.Context, bucketName, objectName string) (*IndexInjector, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create client")
	}
	defer client.Close()
	object := client.Bucket(bucketName).Object(objectName)
	r, err := object.NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create reader for remote index")
	}
	defer r.Close()
	idx, err := index.Read(r)
	if err != nil {
		return nil, err
	}
	return &IndexInjector{idx: idx}, nil
}

// NewInjector returns a new IndexInjector.
func NewInjector(location string) (*IndexInjector, error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	idx, err := index.Read(file)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return &IndexInjector{idx: idx}, nil
}

// Middleware injects the index into the context.
func (s *IndexInjector) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set(InjectorContextKey, s.idx)
		return next(c)
	}
}
