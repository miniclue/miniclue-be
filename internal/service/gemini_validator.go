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
	geminiBaseURL        = "https://generativelanguage.googleapis.com/v1beta"
	geminiModelsEndpoint = "/models"
)

// GeminiValidator validates Gemini API keys by making a test API call
type GeminiValidator interface {
	ValidateAPIKey(ctx context.Context, apiKey string) error
}

type geminiValidator struct {
	client  *http.Client
	baseURL string
}

// NewGeminiValidator creates a new Gemini API key validator
func NewGeminiValidator() GeminiValidator {
	return &geminiValidator{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: geminiBaseURL,
	}
}

// ValidateAPIKey validates a Gemini API key by making a test call to the models endpoint
func (v *geminiValidator) ValidateAPIKey(ctx context.Context, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Create request to Gemini models endpoint with API key as query parameter
	url := fmt.Sprintf("%s%s?key=%s", v.baseURL, geminiModelsEndpoint, apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	// Set content type header
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
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Status  string `json:"status"`
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
				Status  string `json:"status"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return fmt.Errorf("API key validation failed: %s", errorResp.Error.Message)
		}
		return fmt.Errorf("API key validation failed: HTTP %d", resp.StatusCode)
	}

	// Verify response is valid JSON (models list)
	var modelsResp struct {
		Models []interface{} `json:"models"`
	}
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return fmt.Errorf("invalid response format from Gemini: %w", err)
	}

	return nil
}
