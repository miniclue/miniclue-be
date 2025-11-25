package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	openAIBaseURL        = "https://api.openai.com/v1"
	openAIModelsEndpoint = "/models"
	validationTimeout    = 10 * time.Second
)

// OpenAIValidator validates OpenAI API keys by making a test API call
type OpenAIValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string) error
}

type openAIValidator struct {
	client  *http.Client
	baseURL string
}

// NewOpenAIValidator creates a new OpenAI API key validator
func NewOpenAIValidator() OpenAIValidator {
	return &openAIValidator{
		client: &http.Client{
			Timeout: validationTimeout,
		},
		baseURL: openAIBaseURL,
	}
}

// ValidateAPIKey validates an OpenAI API key by making a test call to the models endpoint
func (v *openAIValidator) ValidateAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Create request to OpenAI models endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.baseURL+openAIModelsEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read validation response: %w", err)
	}

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return fmt.Errorf("invalid API key: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("invalid API key: unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return fmt.Errorf("API key validation failed: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("API key validation failed: HTTP %d", resp.StatusCode)
	}

	// Verify response is valid JSON (models list)
	var modelsResp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return fmt.Errorf("invalid response format from OpenAI: %w", err)
	}

	return nil
}
