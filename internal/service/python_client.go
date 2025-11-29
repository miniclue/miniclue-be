package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type PythonClient interface {
	StreamChat(ctx context.Context, lectureID, chatID, userID string, messageParts []map[string]interface{}, model string) (io.ReadCloser, error)
	GenerateChatTitle(ctx context.Context, lectureID, chatID, userID string, userMessageParts []map[string]interface{}, assistantMessageParts []map[string]interface{}) (string, error)
}

type pythonClient struct {
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

func NewPythonClient(baseURL string, logger zerolog.Logger) PythonClient {
	return &pythonClient{
		baseURL: baseURL,
		client:  &http.Client{
			// No timeout for streaming - rely on context cancellation instead
			// This allows long-running streaming responses without premature cancellation
		},
		logger: logger.With().Str("service", "PythonClient").Logger(),
	}
}

type ChatRequest struct {
	LectureID string                   `json:"lecture_id"`
	ChatID    string                   `json:"chat_id"`
	UserID    string                   `json:"user_id"`
	Message   []map[string]interface{} `json:"message"`
	Model     string                   `json:"model"`
}

func (c *pythonClient) StreamChat(ctx context.Context, lectureID, chatID, userID string, messageParts []map[string]interface{}, model string) (io.ReadCloser, error) {
	reqBody := ChatRequest{
		LectureID: lectureID,
		ChatID:    chatID,
		UserID:    userID,
		Message:   messageParts,
		Model:     model,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	url := fmt.Sprintf("%s/chat", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request to Python service: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Read error body for better error messages
		bodyBytes, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if readErr != nil {
			c.logger.Warn().Err(readErr).Int("status_code", resp.StatusCode).Msg("Failed to read error body from Python service")
			return nil, fmt.Errorf("python service returned status %d", resp.StatusCode)
		}

		errorMsg := string(bodyBytes)
		c.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("error_body", errorMsg).
			Msg("Python service returned error")

		return nil, fmt.Errorf("python service returned status %d: %s", resp.StatusCode, errorMsg)
	}

	return resp.Body, nil
}

type TitleRequest struct {
	LectureID        string                   `json:"lecture_id"`
	ChatID           string                   `json:"chat_id"`
	UserID           string                   `json:"user_id"`
	UserMessage      []map[string]interface{} `json:"user_message"`
	AssistantMessage []map[string]interface{} `json:"assistant_message"`
}

type TitleResponse struct {
	Title string `json:"title"`
}

func (c *pythonClient) GenerateChatTitle(ctx context.Context, lectureID, chatID, userID string, userMessageParts []map[string]interface{}, assistantMessageParts []map[string]interface{}) (string, error) {
	reqBody := TitleRequest{
		LectureID:        lectureID,
		ChatID:           chatID,
		UserID:           userID,
		UserMessage:      userMessageParts,
		AssistantMessage: assistantMessageParts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request body: %w", err)
	}

	url := fmt.Sprintf("%s/chat/generate-title", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request to Python service: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn().Err(closeErr).Msg("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			c.logger.Warn().Err(readErr).Int("status_code", resp.StatusCode).Msg("Failed to read error body from Python service")
			return "", fmt.Errorf("python service returned status %d", resp.StatusCode)
		}

		errorMsg := string(bodyBytes)
		c.logger.Error().
			Int("status_code", resp.StatusCode).
			Str("error_body", errorMsg).
			Msg("Python service returned error")

		return "", fmt.Errorf("python service returned status %d: %s", resp.StatusCode, errorMsg)
	}

	var titleResp TitleResponse
	if err := json.NewDecoder(resp.Body).Decode(&titleResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return titleResp.Title, nil
}

// ParseSSEChunk parses a single SSE chunk from the stream.
// SSE format: "data: <json>\n\n" where blank line separates events.
// Handles comments (lines starting with ":") and empty lines.
func ParseSSEChunk(reader *bufio.Reader) (map[string]interface{}, error) {
	var dataLine string
	foundData := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if foundData {
					// We found data but hit EOF before blank line - this is valid
					break
				}
				return nil, io.EOF
			}
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		// Empty line indicates end of SSE event
		if len(line) == 0 {
			if foundData {
				break
			}
			// Skip blank lines before data
			continue
		}

		// Skip comments (SSE spec allows comments starting with ":")
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse data line
		if strings.HasPrefix(line, "data: ") {
			dataLine = line[6:] // Remove "data: " prefix
			foundData = true
			// Continue reading until we hit blank line or EOF
			continue
		}

		// If we already found data and hit a non-empty, non-comment line,
		// this might be malformed SSE, but we'll try to parse what we have
		if foundData {
			break
		}
	}

	if !foundData {
		return nil, fmt.Errorf("no data line found in SSE chunk")
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(dataLine), &result); err != nil {
		return nil, fmt.Errorf("unmarshaling SSE data %q: %w", dataLine, err)
	}

	return result, nil
}
