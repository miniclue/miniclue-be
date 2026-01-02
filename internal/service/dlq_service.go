package service

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"app/internal/api/v1/dto"
	"app/internal/model"
	"app/internal/repository"

	"github.com/rs/zerolog"
)

// DLQService defines the interface for Dead Letter Queue operations.
type DLQService interface {
	ProcessAndSave(ctx context.Context, req *dto.PubSubPushRequest) error
}

// dlqService is the implementation of DLQService.
type dlqService struct {
	repo      repository.DLQRepository
	dlqLogger zerolog.Logger
}

// NewDLQService creates a new DLQService.
func NewDLQService(repo repository.DLQRepository, logger zerolog.Logger) DLQService {
	return &dlqService{
		repo:      repo,
		dlqLogger: logger.With().Str("service", "DLQService").Logger(),
	}
}

// ProcessAndSave processes and saves a message from a Pub/Sub push request.
func (s *dlqService) ProcessAndSave(ctx context.Context, req *dto.PubSubPushRequest) error {
	// Decode the base64-encoded payload
	decodedPayload, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		s.dlqLogger.Warn().Err(err).Str("message_id", req.Message.MessageID).Msg("Failed to decode DLQ message payload, saving as is")
		decodedPayload = []byte(req.Message.Data) // Save the raw base64 string
	}

	// Marshal attributes to JSON bytes, if they exist
	var attributesJSON []byte
	if len(req.Message.Attributes) > 0 {
		var err error
		attributesJSON, err = json.Marshal(req.Message.Attributes)
		if err != nil {
			s.dlqLogger.Warn().Err(err).Str("message_id", req.Message.MessageID).Msg("Failed to marshal DLQ message attributes")
		}
	}

	// Create the model for the database
	dbMessage := &model.DeadLetterMessage{
		SubscriptionName: req.Subscription,
		MessageID:        req.Message.MessageID,
		Payload:          decodedPayload,
		Attributes:       attributesJSON,
		Status:           "unprocessed", // Default status
	}

	// Save to the database
	if err := s.repo.Create(ctx, dbMessage); err != nil {
		s.dlqLogger.Error().Err(err).Str("subscription", dbMessage.SubscriptionName).Msg("Failed to save DLQ message")
		return err
	}
	// Success is expected, no need to log it
	return nil
}
