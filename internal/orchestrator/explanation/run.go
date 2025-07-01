package explanation

import (
	"bytes"
	"context"
	"database/sql"
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

// Run starts the explanation orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	// Load explanation-specific config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config in explanation orchestrator: %v", err)
	}
	queue := cfg.ExplanationQueueName
	dlq := cfg.ExplanationDeadLetterQueueName
	baseURL := strings.TrimRight(cfg.PythonServiceBaseURL, "/")
	explainEndpoint := fmt.Sprintf("%s/explain", baseURL)
	logger.Info().Str("queue", queue).Str("endpoint", explainEndpoint).Msg("Starting explanation orchestrator")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down explanation orchestrator")
			return nil
		default:
		}

		logger.Info().Msg("Reading explanation queue")
		msgs, err := client.ReadWithPoll(ctx, queue, cfg.ExplanationPollTimeoutSec, cfg.ExplanationPollMaxMsg)
		if err != nil {
			logger.Error().Err(err).Msg("Error reading explanation queue")
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]
		logger.Info().Int64("msg_id", msg.ID).Msgf("Received explanation job: %s", string(msg.Data))

		// Parse payload
		var payload struct {
			SlideID     string `json:"slide_id"`
			LectureID   string `json:"lecture_id"`
			SlideNumber int    `json:"slide_number"`
		}
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal explanation payload; deleting message")
			client.Delete(ctx, queue, []int64{msg.ID})
			continue
		}

		// Wait for previous explanation if not the first slide
		if payload.SlideNumber > 1 {
			var exists int
			err := client.QueryRow(ctx,
				"SELECT 1 FROM explanations WHERE lecture_id=$1 AND slide_number=$2",
				payload.LectureID, payload.SlideNumber-1,
			).Scan(&exists)
			if err == sql.ErrNoRows {
				logger.Info().Int("slide_number", payload.SlideNumber).Msg("Previous explanation not ready; retrying")
				time.Sleep(time.Second)
				continue
			} else if err != nil {
				logger.Error().Err(err).Msg("Error checking previous explanation; retrying")
				time.Sleep(time.Second)
				continue
			}
		}

		// Call Python explanation service with retry/backoff
		backoff := time.Duration(cfg.ExplanationBackoffInitialSec) * time.Second
		var httpErr error
		for attempt := 1; attempt <= cfg.ExplanationMaxRetries; attempt++ {
			requestTimeout := time.Duration(cfg.ExplanationRequestTimeoutSec) * time.Second
			ctxReq, cancel := context.WithTimeout(ctx, requestTimeout)
			reqBody, _ := json.Marshal(payload)
			req, _ := http.NewRequestWithContext(ctxReq, http.MethodPost, explainEndpoint, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			duration := time.Since(start)
			cancel()

			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				logger.Info().Str("duration", duration.String()).Msg("Explanation service succeeded")
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
			logger.Error().Err(httpErr).Int("attempt", attempt).Msg("Explanation service call failed, retrying")
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Duration(cfg.ExplanationBackoffMaxSec)*time.Second {
				backoff = time.Duration(cfg.ExplanationBackoffMaxSec) * time.Second
			}
		}

		// Handle failure after retries
		if httpErr != nil {
			errorDetails := map[string]string{"stage": "explanation", "message": httpErr.Error()}
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
				logger.Error().Err(err).Msg("Error deleting explanation message after failure")
			}
			logger.Warn().Int("attempts", cfg.ExplanationMaxRetries).Str("lecture_id", payload.LectureID).Err(httpErr).Msg("Exhausted all explanation retries; moving job to DLQ")
			continue
		}

		// Acknowledge message on success
		if err := client.Delete(ctx, queue, []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting explanation message")
		}
	}
}
