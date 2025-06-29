package ingestion

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

// Run starts the ingestion orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	// Load ingestion-specific config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config in ingestion orchestrator: %v", err)
	}
	queue := cfg.IngestionQueueName
	// Build the Python ingestion endpoint from base URL
	baseURL := strings.TrimRight(cfg.PythonServiceBaseURL, "/")
	ingestEndpoint := fmt.Sprintf("%s/ingest", baseURL)
	logger.Info().Str("queue", queue).Str("endpoint", ingestEndpoint).Msg("Starting ingestion orchestrator")
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down ingestion orchestrator")
			return nil
		default:
		}
		logger.Info().Msg("Reading ingestion queue")
		// Read one message from the ingestion queue
		msgs, err := client.ReadWithPoll(ctx, queue, cfg.IngestionPollTimeoutSec, cfg.IngestionPollMaxMsg)
		if err != nil {
			logger.Error().Err(err).Msg("Error reading ingestion queue")
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]
		logger.Info().Int64("msg_id", msg.ID).Msgf("Received ingestion job: %s", string(msg.Data))

		// Parse payload
		var payload struct {
			LectureID   string `json:"lecture_id"`
			StoragePath string `json:"storage_path"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal ingestion payload; deleting message")
			client.Delete(ctx, queue, []int64{msg.ID})
			continue
		}

		// Update lecture status to parsing
		if err := client.Exec(ctx, "UPDATE lectures SET status=$1 WHERE id=$2", "parsing", payload.LectureID); err != nil {
			logger.Error().Err(err).Str("lecture_id", payload.LectureID).Msg("Failed to update lecture status to parsing; will retry")
			time.Sleep(time.Second)
			continue
		}

		// Call Python ingestion service with retry/backoff
		backoff := time.Duration(cfg.IngestionBackoffInitialSec) * time.Second
		var httpErr error
		for attempt := 1; attempt <= cfg.IngestionMaxRetries; attempt++ {
			ctxReq, cancel := context.WithTimeout(ctx, 10*time.Second)
			reqBody, _ := json.Marshal(payload)
			// Exponential backoff retry against Python ingestion service
			req, _ := http.NewRequestWithContext(ctxReq, http.MethodPost, ingestEndpoint, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)
			cancel()

			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				logger.Info().Str("duration", duration.String()).Msg("Ingestion service succeeded")
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
			logger.Error().Err(httpErr).Int("attempt", attempt).Msg("Ingestion service call failed, retrying")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Duration(cfg.IngestionBackoffMaxSec)*time.Second {
				backoff = time.Duration(cfg.IngestionBackoffMaxSec) * time.Second
			}
		}
		if httpErr != nil {
			// Mark lecture as failed, recording structured error details
			errorDetails := map[string]string{
				"stage":   "ingestion",
				"message": httpErr.Error(),
			}
			detailsBytes, _ := json.Marshal(errorDetails)

			updateQuery := "UPDATE lectures SET status=$1, error_details=$2 WHERE id=$3"
			if err := client.Exec(ctx, updateQuery, "failed", detailsBytes, payload.LectureID); err != nil {
				logger.Error().Err(err).Str("lecture_id", payload.LectureID).Msg("Failed to update lecture status to failed")
			}
			// Send the failed job to dead-letter queue
			dlq := cfg.IngestionDeadLetterQueueName
			if msgBytes, err := json.Marshal(payload); err == nil {
				if err := client.Send(ctx, dlq, msgBytes); err != nil {
					logger.Error().Err(err).Str("dlq", dlq).Msg("Failed to send message to dead-letter queue")
				}
			} else {
				logger.Error().Err(err).Msg("Failed to marshal payload for dead-letter queue")
			}
			// Acknowledge (delete) the original message so it won't retry
			if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
				logger.Error().Err(err).Msg("Error deleting ingestion message after failure")
			}
			logger.Warn().
				Int("attempts", cfg.IngestionMaxRetries).
				Str("lecture_id", payload.LectureID).
				Err(httpErr).
				Msg("Exhausted all ingestion retries; moving job to DLQ")
			continue
		}

		// Acknowledge message
		if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting ingestion message")
		}

		// Update lecture status to embedding
		if err := client.Exec(ctx, "UPDATE lectures SET status=$1 WHERE id=$2", "embedding", payload.LectureID); err != nil {
			logger.Error().Err(err).Str("lecture_id", payload.LectureID).Msg("Failed to update lecture status to embedding")
		}
	}
}
