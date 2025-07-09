package pubsub

import (
	"context"
	"fmt"

	"app/internal/config"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

// Publisher defines an interface for publishing messages.
type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) (string, error)
}

// PubSubPublisher is an implementation of Publisher using Google Pub/Sub.
type PubSubPublisher struct {
	client *pubsub.Client
}

// NewPublisher creates a new PubSubPublisher using the GCP project from config.
func NewPublisher(ctx context.Context, cfg *config.Config) (*PubSubPublisher, error) {
	var opts []option.ClientOption
	if cfg.PubSubEmulatorHost != "" {
		opts = append(opts, option.WithEndpoint(cfg.PubSubEmulatorHost), option.WithoutAuthentication())
	}

	client, err := pubsub.NewClient(ctx, cfg.GCPProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}
	return &PubSubPublisher{client: client}, nil
}

// Publish sends the payload to the given Pub/Sub topic and returns the message ID.
func (p *PubSubPublisher) Publish(ctx context.Context, topic string, payload []byte) (string, error) {
	t := p.client.Topic(topic)
	result := t.Publish(ctx, &pubsub.Message{Data: payload})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}
	return id, nil
}
