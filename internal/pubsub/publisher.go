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
	var projectID string

	// Determine environment based on PubSubEmulatorHost presence
	if cfg.PubSubEmulatorHost != "" {
		// Local environment
		projectID = cfg.GCPProjectIDLocal
		opts = append(opts, option.WithEndpoint(cfg.PubSubEmulatorHost), option.WithoutAuthentication())
	} else {
		// Production/Staging environment - use staging by default
		// In production, you would typically set the appropriate project ID via environment variables
		if cfg.GCPProjectIDStaging != "" {
			projectID = cfg.GCPProjectIDStaging
		} else if cfg.GCPProjectIDProd != "" {
			projectID = cfg.GCPProjectIDProd
		} else {
			return nil, fmt.Errorf("no GCP Project ID configured for staging or production")
		}
	}

	if projectID == "" {
		return nil, fmt.Errorf("GCP Project ID is not set for the current environment")
	}

	client, err := pubsub.NewClient(ctx, projectID, opts...)
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
