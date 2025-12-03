package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicBaseURL        = "https://api.anthropic.com/v1"
	anthropicModelsEndpoint = "/messages"
)

// AnthropicValidator validates Anthropic API keys by making a test API call
type AnthropicValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string) error
}

type anthropicValidator struct {
	client  *http.Client
	baseURL string
}

// NewAnthropicValidator creates a new Anthropic API key validator
func NewAnthropicValidator() AnthropicValidator {
	return &anthropicValidator{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: anthropicBaseURL,
	}
}

// ValidateAPIKey validates an Anthropic API key by making a test call to the messages endpoint
func (v *anthropicValidator) ValidateAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Create a minimal test request body
	requestBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request to Anthropic messages endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.baseURL+anthropicModelsEndpoint, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	// Set headers
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
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

	return nil
}
