package main

import (
	"context"
	"fmt"
	"time"

	"app/internal/config"
	"app/internal/logger"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// The base URL for the Python services that will receive push messages.
// On Docker for Mac/Windows, 'host.docker.internal' lets containers reach the host machine.
const pythonAPIBASEURL = "http://host.docker.internal:8000"

// The URL for the Go API gateway that will receive dead-letter messages.
const gatewayAPIURL = "http://host.docker.internal:8080/v1/dlq"

func main() {
	logger := logger.New()

	// Load environment variables from a .env file if it exists.
	if err := godotenv.Load(); err != nil {
		logger.Warn().Msg("No .env file found, relying on system environment variables.")
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Failed to load config: %v", err)
	}

	if cfg.PubSubEmulatorHost == "" {
		logger.Fatal().Msg("PUBSUB_EMULATOR_HOST is not set. Please configure it in your .env file.")
	}

	client, err := pubsub.NewClient(ctx, cfg.GCPProjectID,
		option.WithEndpoint(cfg.PubSubEmulatorHost),
		option.WithoutAuthentication(),
	)
	if err != nil {
		logger.Fatal().Msgf("Failed to create Pub/Sub client: %v", err)
	}
	defer client.Close()

	// --- Deletion Phase ---
	logger.Info().Msg("\n--- Deleting all existing subscriptions and topics for a clean setup ---")

	// Delete all subscriptions
	subs := client.Subscriptions(ctx)
	for {
		sub, err := subs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Fatal().Msgf("Failed to list subscriptions: %v", err)
		}
		logger.Info().Msgf("Deleting subscription: %s", sub.ID())
		if err := sub.Delete(ctx); err != nil {
			logger.Warn().Msgf("ERROR: Failed to delete subscription %s: %v", sub.ID(), err)
		}
	}

	// Delete all topics
	topicsToDelete := client.Topics(ctx)
	for {
		topic, err := topicsToDelete.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Fatal().Msgf("Failed to list topics: %v", err)
		}
		logger.Info().Msgf("Deleting topic: %s", topic.ID())
		if err := topic.Delete(ctx); err != nil {
			logger.Warn().Msgf("ERROR: Failed to delete topic %s: %v", topic.ID(), err)
		}
	}

	logger.Info().Msg("\n--- Deletion complete. Starting creation phase. ---")

	// --- Creation Phase ---
	topicsToCreate := []string{"ingestion", "embedding", "explanation", "summary"}

	for _, topicID := range topicsToCreate {
		dlqTopicID := topicID + "-dlq"
		subID := topicID + "-sub"
		pushEndpoint := fmt.Sprintf("%s/%s", pythonAPIBASEURL, topicID)

		logger.Info().Msgf("\n--- Processing topic: %s ---", topicID)

		// 1. Create the Dead-Letter Queue (DLQ) Topic
		logger.Info().Msgf("Creating DLQ topic: %s", dlqTopicID)
		dlqTopic, err := client.CreateTopic(ctx, dlqTopicID)
		if err != nil {
			logger.Fatal().Msgf("Failed to create DLQ topic '%s': %v", dlqTopicID, err)
		}

		// 2. Create the Main Topic
		logger.Info().Msgf("Creating main topic: %s", topicID)
		mainTopic, err := client.CreateTopic(ctx, topicID)
		if err != nil {
			logger.Fatal().Msgf("Failed to create topic '%s': %v", topicID, err)
		}

		// 3. Create the Push Subscription
		logger.Info().Msgf("Creating push subscription: %s", subID)
		subConfig := pubsub.SubscriptionConfig{
			Topic:       mainTopic,
			PushConfig:  pubsub.PushConfig{Endpoint: pushEndpoint},
			AckDeadline: 60 * time.Second,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: 10 * time.Second,
				MaximumBackoff: 600 * time.Second,
			},
			DeadLetterPolicy: &pubsub.DeadLetterPolicy{
				DeadLetterTopic:     dlqTopic.String(),
				MaxDeliveryAttempts: 5,
			},
		}
		if _, err := client.CreateSubscription(ctx, subID, subConfig); err != nil {
			logger.Fatal().Msgf("Failed to create subscription '%s': %v", subID, err)
		}
	}

	// --- Create DLQ Subscriptions ---
	logger.Info().Msg("\n--- Creating DLQ Subscriptions ---")
	for _, topicID := range topicsToCreate {
		dlqTopicID := topicID + "-dlq"
		dlqSubID := dlqTopicID + "-sub"
		dlqTopic := client.Topic(dlqTopicID)

		logger.Info().Msgf("Creating DLQ subscription: %s for topic %s", dlqSubID, dlqTopicID)
		subConfig := pubsub.SubscriptionConfig{
			Topic:       dlqTopic,
			PushConfig:  pubsub.PushConfig{Endpoint: gatewayAPIURL},
			AckDeadline: 60 * time.Second,
		}
		if _, err := client.CreateSubscription(ctx, dlqSubID, subConfig); err != nil {
			logger.Fatal().Msgf("Failed to create DLQ subscription '%s': %v", dlqSubID, err)
		}
	}

	logger.Info().Msg("\nPub/Sub local setup complete.")
}
