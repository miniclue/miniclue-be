package service

import (
	"context"
	"fmt"

	"app/internal/config"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

type SecretManagerService interface {
	StoreUserAPIKey(ctx context.Context, userID, provider, apiKey string) error
	GetUserAPIKey(ctx context.Context, userID, provider string) (string, error)
	DeleteUserAPIKey(ctx context.Context, userID, provider string) error
}

type secretManagerService struct {
	client    *secretmanager.Client
	projectID string
}

func NewSecretManagerService(ctx context.Context, cfg *config.Config) (SecretManagerService, error) {
	projectID := cfg.GetGCPProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("GCP Project ID is not set for the current environment")
	}

	var opts []option.ClientOption
	// Note: Secret Manager requires a real GCP project even for local development.
	// Set GCP_PROJECT_ID_LOCAL to your local development GCP project ID.

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}

	return &secretManagerService{
		client:    client,
		projectID: projectID,
	}, nil
}

func (s *secretManagerService) StoreUserAPIKey(ctx context.Context, userID, provider, apiKey string) error {
	secretName := fmt.Sprintf("user-%s-%s-key", userID, provider)
	secretPath := fmt.Sprintf("projects/%s/secrets/%s", s.projectID, secretName)

	secretExists := true
	_, err := s.client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: secretPath,
	})
	if err != nil {
		secretExists = false
	}

	if !secretExists {
		createReq := &secretmanagerpb.CreateSecretRequest{
			Parent:   fmt.Sprintf("projects/%s", s.projectID),
			SecretId: secretName,
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
			},
		}
		_, err := s.client.CreateSecret(ctx, createReq)
		if err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretPath,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(apiKey),
		},
	}

	_, err = s.client.AddSecretVersion(ctx, addVersionReq)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	return nil
}

func (s *secretManagerService) GetUserAPIKey(ctx context.Context, userID, provider string) (string, error) {
	secretName := fmt.Sprintf("user-%s-%s-key", userID, provider)
	resourceName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", s.projectID, secretName)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	}

	result, err := s.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

func (s *secretManagerService) DeleteUserAPIKey(ctx context.Context, userID, provider string) error {
	secretName := fmt.Sprintf("user-%s-%s-key", userID, provider)
	secretPath := fmt.Sprintf("projects/%s/secrets/%s", s.projectID, secretName)

	req := &secretmanagerpb.DeleteSecretRequest{
		Name: secretPath,
	}

	err := s.client.DeleteSecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
