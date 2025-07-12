package service

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"app/internal/api/v1/dto"
	"app/internal/model"
	"app/internal/repository"
)

type DLQService interface {
	ProcessAndSave(ctx context.Context, req *dto.PubSubPushRequest) error
}

type dlqService struct {
	repo repository.DLQRepository
}

func NewDLQService(repo repository.DLQRepository) DLQService {
	return &dlqService{repo: repo}
}

func (s *dlqService) ProcessAndSave(ctx context.Context, req *dto.PubSubPushRequest) error {
	// Decode the base64-encoded payload
	decodedPayload, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		// If we can't even decode it, save the raw data
		decodedPayload = []byte(req.Message.Data)
	}

	// Marshal attributes to a JSON string, if they exist
	var attributesJSON *string
	if len(req.Message.Attributes) > 0 {
		attrBytes, err := json.Marshal(req.Message.Attributes)
		if err == nil {
			attrStr := string(attrBytes)
			attributesJSON = &attrStr
		}
	}

	// Create the model for the database
	dbMessage := &model.DeadLetterMessage{
		SubscriptionName: req.Subscription,
		MessageID:        req.Message.MessageID,
		Payload:          string(decodedPayload),
		Attributes:       attributesJSON,
		Status:           "unprocessed", // Default status
	}

	// Save to the database
	return s.repo.Create(ctx, dbMessage)
}
