package main

import (
	"context"
	"fmt"
	"time"

	"app/internal/config"
	"app/internal/logger"

	"cloud.google.com/go/pubsub"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// For local development, 'host.docker.internal' lets containers reach the host machine.
const pythonAPIBaseURLLocal = "http://host.docker.internal:8000"
const gatewayAPIURLLocal = "http://host.docker.internal:8080/v1/dlq"

func main() {
	// Load environment variables early for local development
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, relying on system environment variables.")
	}

	// Configuration Setup for Local Environment
	logger := logger.New()
	logger.Info().Msg("Starting Pub/Sub setup for the local environment.")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Failed to load config: %v", err)
	}

	projectID := cfg.GCPProjectIDLocal
	if projectID == "" {
		logger.Fatal().Msg("GCP_PROJECT_ID_LOCAL is not set in the environment.")
	}
	if cfg.PubSubEmulatorHost == "" {
		logger.Fatal().Msg("PUBSUB_EMULATOR_HOST must be set for local environment.")
	}

	clientOptions := []option.ClientOption{
		option.WithEndpoint(cfg.PubSubEmulatorHost),
		option.WithoutAuthentication(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, projectID, clientOptions...)
	if err != nil {
		logger.Fatal().Msgf("Failed to create Pub/Sub client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Error().Msgf("Failed to close pubsub client: %v", err)
		}
	}()

	// Execute Local Reset and Creation
	resetLocalEmulator(ctx, client, logger)
	createOrUpdateResources(ctx, client, logger)

	logger.Info().Msg("\nPub/Sub setup for local environment complete.")
}

// resetLocalEmulator performs a destructive reset of all topics and subscriptions.
// This should ONLY be used against the local emulator.
func resetLocalEmulator(ctx context.Context, client *pubsub.Client, logger zerolog.Logger) {
	logger.Info().Msg("\n--- Deleting all existing resources for a clean local setup ---")

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
}

// createOrUpdateResources ensures all topics and subscriptions exist and are correctly configured for the local environment.
func createOrUpdateResources(ctx context.Context, client *pubsub.Client, logger zerolog.Logger) {
	topics := []string{"ingestion", "embedding", "image-analysis"}
	sevenDays := 7 * 24 * time.Hour

	for _, topicID := range topics {
		// Name Generation for Local Env
		dlqTopicID := topicID + "-dlq"
		subID := fmt.Sprintf("%s-sub", topicID)
		dlqSubID := fmt.Sprintf("%s-dlq-sub", topicID)

		// Endpoint Generation for Local Env
		pushEndpoint := fmt.Sprintf("%s/%s", pythonAPIBaseURLLocal, topicID)

		logger.Info().Msgf("\n--- Ensuring resources for topic: %s (Local Env) ---", topicID)
		logger.Info().Msgf("Subscription name: %s", subID)
		logger.Info().Msgf("DLQ Subscription name: %s", dlqSubID)

		// Create DLQ and Main topics if they don't exist
		dlqTopic, _ := createTopicIfNotExists(ctx, client, logger, dlqTopicID, sevenDays)
		mainTopic, _ := createTopicIfNotExists(ctx, client, logger, topicID, sevenDays)

		// Subscription Configurations
		mainSubConfig := pubsub.SubscriptionConfig{
			Topic:            mainTopic,
			PushConfig:       pubsub.PushConfig{Endpoint: pushEndpoint},
			AckDeadline:      180 * time.Second,
			ExpirationPolicy: 31 * 24 * time.Hour,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: 10 * time.Second,
				MaximumBackoff: 600 * time.Second,
			},
			DeadLetterPolicy: &pubsub.DeadLetterPolicy{
				DeadLetterTopic:     dlqTopic.String(),
				MaxDeliveryAttempts: 5,
			},
		}
		createOrUpdateSubscription(ctx, client, logger, subID, mainSubConfig)

		dlqSubConfig := pubsub.SubscriptionConfig{
			Topic:            dlqTopic,
			PushConfig:       pubsub.PushConfig{Endpoint: gatewayAPIURLLocal},
			AckDeadline:      180 * time.Second,
			ExpirationPolicy: 31 * 24 * time.Hour,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: 10 * time.Second,
				MaximumBackoff: 600 * time.Second,
			},
		}
		createOrUpdateSubscription(ctx, client, logger, dlqSubID, dlqSubConfig)
	}
}

func createTopicIfNotExists(ctx context.Context, client *pubsub.Client, logger zerolog.Logger, topicID string, retention time.Duration) (*pubsub.Topic, error) {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		logger.Fatal().Msgf("Failed to check if topic %s exists: %v", topicID, err)
	}

	if !exists {
		logger.Info().Msgf("Creating topic: %s with %v retention", topicID, retention)
		return client.CreateTopicWithConfig(ctx, topicID, &pubsub.TopicConfig{
			RetentionDuration: retention,
		})
	}

	logger.Info().Msgf("Topic %s already exists. Verifying settings...", topicID)
	cfg, err := topic.Config(ctx)
	if err != nil {
		logger.Fatal().Msgf("Failed to get config for topic %s: %v", topicID, err)
	}

	if cfg.RetentionDuration != retention {
		logger.Warn().Msgf("  -> WARNING: Mismatched retention for topic %s. Expected %v, but found %v. Please update manually in the GCP console.", topicID, retention, cfg.RetentionDuration)
	} else {
		logger.Info().Msgf("  -> Retention for topic %s is correct.", topicID)
	}
	return topic, nil
}

func createOrUpdateSubscription(ctx context.Context, client *pubsub.Client, logger zerolog.Logger, subID string, config pubsub.SubscriptionConfig) {
	sub := client.Subscription(subID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		logger.Fatal().Msgf("Failed to check if subscription %s exists: %v", subID, err)
	}

	if !exists {
		logger.Info().Msgf("Creating subscription %s with endpoint %s", subID, config.PushConfig.Endpoint)
		if _, err := client.CreateSubscription(ctx, subID, config); err != nil {
			logger.Fatal().Msgf("Failed to create subscription '%s': %v", subID, err)
		}
		return
	}

	logger.Info().Msgf("Subscription %s already exists. Checking configuration...", subID)
	existingConfig, err := sub.Config(ctx)
	if err != nil {
		logger.Fatal().Msgf("Failed to get config for subscription '%s': %v", subID, err)
	}

	updateRequired := false
	updateCfg := pubsub.SubscriptionConfigToUpdate{}

	if existingConfig.PushConfig.Endpoint != config.PushConfig.Endpoint {
		logger.Info().Msgf("  -> Mismatch found for PushConfig.Endpoint on subscription '%s'.", subID)
		updateRequired = true
		updateCfg.PushConfig = &config.PushConfig
	}

	if existingConfig.AckDeadline != config.AckDeadline {
		logger.Info().Msgf("  -> Mismatch found for AckDeadline on subscription '%s'.", subID)
		updateRequired = true
		updateCfg.AckDeadline = config.AckDeadline
	}

	// Compare RetryPolicy
	if (existingConfig.RetryPolicy == nil && config.RetryPolicy != nil) || (existingConfig.RetryPolicy != nil && config.RetryPolicy == nil) {
		updateRequired = true
	} else if existingConfig.RetryPolicy != nil && config.RetryPolicy != nil {
		if existingConfig.RetryPolicy.MinimumBackoff != config.RetryPolicy.MinimumBackoff || existingConfig.RetryPolicy.MaximumBackoff != config.RetryPolicy.MaximumBackoff {
			updateRequired = true
		}
	}

	if updateRequired {
		logger.Info().Msgf("  -> Updating subscription '%s'", subID)
		// For updates, we need to explicitly set all fields we want to change.
		if updateCfg.PushConfig == nil {
			updateCfg.PushConfig = &existingConfig.PushConfig
		}
		if updateCfg.AckDeadline == 0 {
			updateCfg.AckDeadline = existingConfig.AckDeadline
		}
		updateCfg.RetryPolicy = config.RetryPolicy // Always set the desired retry policy

		if _, err := sub.Update(ctx, updateCfg); err != nil {
			logger.Fatal().Msgf("Failed to update subscription '%s': %v", subID, err)
		}
	} else {
		logger.Info().Msg("  -> Configuration is up to date.")
	}
}
