package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"app/internal/config"
	"app/internal/pgmq"

	"github.com/rs/zerolog"
)

// Run starts the embedding orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	// Load embedding-specific config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config in embedding orchestrator: %v", err)
	}
	queue := cfg.EmbeddingQueueName
	dlq := cfg.EmbeddingDeadLetterQueueName
	baseURL := strings.TrimRight(cfg.PythonServiceBaseURL, "/")
	embedEndpoint := fmt.Sprintf("%s/embed", baseURL)
	logger.Info().Str("queue", queue).Str("endpoint", embedEndpoint).Msg("Starting embedding orchestrator")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down embedding orchestrator")
			return nil
		default:
		}
		logger.Info().Msg("Reading embedding queue")
		// Read one message from the embedding queue
		msgs, err := client.ReadWithPoll(ctx, queue, cfg.EmbeddingPollTimeoutSec, cfg.EmbeddingPollMaxMsg)
		if err != nil {
			logger.Error().Err(err).Msg("Error reading embedding queue")
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]
		logger.Info().Int64("msg_id", msg.ID).Msgf("Received embedding job: %s", string(msg.Data))

		var payload struct {
			ChunkID     string `json:"chunk_id"`
			SlideID     string `json:"slide_id"`
			LectureID   string `json:"lecture_id"`
			SlideNumber int    `json:"slide_number"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal embedding payload; deleting message")
			client.Delete(ctx, queue, []int64{msg.ID})
			continue
		}

		backoff := time.Duration(cfg.EmbeddingBackoffInitialSec) * time.Second
		var httpErr error
		for attempt := 1; attempt <= cfg.EmbeddingMaxRetries; attempt++ {
			ctxReq, cancel := context.WithTimeout(ctx, 10*time.Second)
			reqBody, _ := json.Marshal(payload)
			req, _ := http.NewRequestWithContext(ctxReq, http.MethodPost, embedEndpoint, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)
			cancel()

			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				logger.Info().Str("duration", duration.String()).Msg("Embedding service succeeded")
				httpErr = nil
				break
			}
			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				httpErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
			} else {
				httpErr = err
			}
			logger.Error().Err(httpErr).Int("attempt", attempt).Msg("Embedding service call failed, retrying")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Duration(cfg.EmbeddingBackoffMaxSec)*time.Second {
				backoff = time.Duration(cfg.EmbeddingBackoffMaxSec) * time.Second
			}
		}

		if httpErr != nil {
			errorDetails := map[string]string{"stage": "embedding", "message": httpErr.Error()}
			detailsBytes, _ := json.Marshal(errorDetails)

			if err := client.Exec(ctx, "UPDATE lectures SET status=$1, error_details=$2 WHERE id=$3", "failed", detailsBytes, payload.LectureID); err != nil {
				logger.Error().Err(err).Str("lecture_id", payload.LectureID).Msg("Failed to update lecture status to failed")
			}

			if payloadBytes, err := json.Marshal(payload); err == nil {
				if err := client.Send(ctx, dlq, payloadBytes); err != nil {
					logger.Error().Err(err).Str("dlq", dlq).Msg("Failed to send message to dead-letter queue")
				}
			} else {
				logger.Error().Err(err).Msg("Failed to marshal payload for dead-letter queue")
			}

			if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
				logger.Error().Err(err).Msg("Error deleting embedding message after failure")
			}
			logger.Warn().Int("attempts", cfg.EmbeddingMaxRetries).Str("lecture_id", payload.LectureID).Err(httpErr).Msg("Exhausted all embedding retries; moving job to DLQ")
			continue
		}

		if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting embedding message")
		}
	}
}
