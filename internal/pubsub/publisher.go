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

	switch cfg.AppEnv {
	case "local":
		projectID = cfg.GCPProjectIDLocal
		if cfg.PubSubEmulatorHost != "" {
			opts = append(opts, option.WithEndpoint(cfg.PubSubEmulatorHost), option.WithoutAuthentication())
		}
	case "stg":
		projectID = cfg.GCPProjectIDStaging
	case "prod":
		projectID = cfg.GCPProjectIDProd
	default:
		return nil, fmt.Errorf("invalid APP_ENV specified: '%s'", cfg.AppEnv)
	}

	if projectID == "" {
		return nil, fmt.Errorf("GCP Project ID for environment '%s' is not set", cfg.AppEnv)
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
