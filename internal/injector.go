package internal

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
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
	sub *pubsub.Subscription
	idx *index.Index
	mu  sync.Mutex
}

func NewCloudInjector(ctx context.Context, bucketName, objectName, projectId, topicName string) (*IndexInjector, error) {
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
	subname := fmt.Sprintf("searchersub-%d", rand.Intn(1000))
	log.Printf("Creating a new subscriber with name: %s", subname)
	pubsubClient, err := pubsub.NewClient(ctx, projectId)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create pubsub client")
	}
	topic := pubsubClient.Topic(topicName)
	sub, err := pubsubClient.CreateSubscription(ctx, subname,
		pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create a new subscription")
	}
	ii := &IndexInjector{sub: sub, idx: idx}
	go ii.start(ctx, bucketName, objectName)
	return ii, nil
}

func (s *IndexInjector) start(ctx context.Context, bucketName, objectName string) {
	err := s.sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		if m.Attributes["eventType"] == "OBJECT_FINALIZE" && m.Attributes["bucketId"] == "setlist-searcher-index" && m.Attributes["objectId"] == "index.txt" {
			client, err := storage.NewClient(ctx)
			if err != nil {
				log.Printf("Failed to create client: %v\n", err)
			}
			defer client.Close()
			object := client.Bucket(bucketName).Object(objectName)
			r, err := object.NewReader(ctx)
			if err != nil {
				log.Printf("failed to create reader for remote index: %v\n", err)
			}
			defer r.Close()
			idx, err := index.Read(r)
			if err != nil {
				log.Printf("Unable to read index: %v\n", err)
			} else {
				s.mu.Lock()
				s.idx = idx
				s.mu.Unlock()
			}
		}
		m.Ack()
	})
	if err != context.Canceled {
		log.Printf("Error handling pubsub notification: %v\n", err)
	}
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
		s.mu.Lock()
		defer s.mu.Unlock()
		c.Set(InjectorContextKey, s.idx)
		return next(c)
	}
}
