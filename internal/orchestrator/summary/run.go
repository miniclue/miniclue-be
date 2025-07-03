package summary

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

// Run starts the summary orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	// Load summary-specific config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config in summary orchestrator: %v", err)
	}
	queue := cfg.SummaryQueueName
	dlq := cfg.SummaryDeadLetterQueueName
	baseURL := strings.TrimRight(cfg.PythonServiceBaseURL, "/")
	summaryEndpoint := fmt.Sprintf("%s/summarize", baseURL)
	logger.Info().Str("queue", queue).Str("endpoint", summaryEndpoint).Msg("Starting summary orchestrator")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down summary orchestrator")
			return nil
		default:
		}

		logger.Info().Msg("Reading summary queue")
		msgs, err := client.ReadWithPoll(ctx, queue, cfg.SummaryPollTimeoutSec, cfg.SummaryPollMaxMsg)
		if err != nil {
			logger.Error().Err(err).Msg("Error reading summary queue")
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]
		logger.Info().Int64("msg_id", msg.ID).Msgf("Received summary job: %s", string(msg.Data))

		// Parse payload
		var payload struct {
			LectureID string `json:"lecture_id"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal summary payload; deleting message")
			client.Delete(ctx, queue, []int64{msg.ID})
			continue
		}

		// Call Python summary service with retry/backoff
		backoff := time.Duration(cfg.SummaryBackoffInitialSec) * time.Second
		var httpErr error
		for attempt := 1; attempt <= cfg.SummaryMaxRetries; attempt++ {
			requestTimeout := time.Duration(cfg.SummaryRequestTimeoutSec) * time.Second
			ctxReq, cancel := context.WithTimeout(ctx, requestTimeout)
			reqBody, _ := json.Marshal(payload)
			req, _ := http.NewRequestWithContext(ctxReq, http.MethodPost, summaryEndpoint, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)
			cancel()

			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				logger.Info().Str("duration", duration.String()).Msg("Summary service succeeded")
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
			logger.Error().Err(httpErr).Int("attempt", attempt).Msg("Summary service call failed, retrying")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Duration(cfg.SummaryBackoffMaxSec)*time.Second {
				backoff = time.Duration(cfg.SummaryBackoffMaxSec) * time.Second
			}
		}

		// Handle failure after retries
		if httpErr != nil {
			errorDetails := map[string]string{"stage": "summary", "message": httpErr.Error()}
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
				logger.Error().Err(err).Msg("Error deleting summary message after failure")
			}
			logger.Warn().Int("attempts", cfg.SummaryMaxRetries).Str("lecture_id", payload.LectureID).Err(httpErr).Msg("Exhausted all summary retries; moving job to DLQ")
			continue
		}

		// Acknowledge message on success
		if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting summary message")
		}
	}
}
