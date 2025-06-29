package summary

import (
	"context"
	"time"

	"app/internal/pgmq"

	"github.com/rs/zerolog"
)

// Run starts the summary orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	logger.Info().Msg("Starting summary orchestrator")
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down summary orchestrator")
			return nil
		default:
		}

		msgs, err := client.ReadWithPoll(ctx, "summary_queue", 30, 1)
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
		// Placeholder for processing...
		if err := client.Delete(ctx, "summary_queue", []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting summary message")
		}
	}
}
