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
	"time"

	"github.com/rs/zerolog"
)

type PythonClient interface {
	StreamChat(ctx context.Context, lectureID, chatID, userID string, messageParts []map[string]interface{}, model string) (io.ReadCloser, error)
}

type pythonClient struct {
	baseURL string
	client  *http.Client
	logger  zerolog.Logger
}

func NewPythonClient(baseURL string, logger zerolog.Logger) PythonClient {
	return &pythonClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for streaming
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
		_ = resp.Body.Close()
		return nil, fmt.Errorf("python service returned status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// ParseSSEChunk parses a single SSE chunk from the stream
func ParseSSEChunk(reader *bufio.Reader) (map[string]interface{}, error) {
	var dataLine string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF
			}
			return nil, err
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataLine = line[6:] // Remove "data: " prefix
			break
		}
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(dataLine), &result); err != nil {
		return nil, fmt.Errorf("unmarshaling SSE data: %w", err)
	}

	return result, nil
}
