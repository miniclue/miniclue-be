package embedding

import (
	"context"
	"time"

	"app/internal/pgmq"

	"github.com/rs/zerolog"
)

// Run starts the embedding orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	logger.Info().Msg("Starting embedding orchestrator")
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down embedding orchestrator")
			return nil
		default:
		}

		msgs, err := client.ReadWithPoll(ctx, "embedding_queue", 30, 1)
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
		// Placeholder for processing...
		if err := client.Delete(ctx, "embedding_queue", []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting embedding message")
		}
	}
}
