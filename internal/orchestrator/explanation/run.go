package explanation

import (
	"context"
	"time"

	"app/internal/pgmq"

	"github.com/rs/zerolog"
)

// Run starts the explanation orchestrator.
func Run(ctx context.Context, logger zerolog.Logger, client *pgmq.Client) error {
	logger.Info().Msg("Starting explanation orchestrator")
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Shutting down explanation orchestrator")
			return nil
		default:
		}

		msgs, err := client.ReadWithPoll(ctx, "explanation_queue", 30, 1)
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
		// Placeholder for processing...
		if err := client.Delete(ctx, "explanation_queue", []int64{msg.ID}); err != nil {
			logger.Error().Err(err).Msg("Error deleting explanation message")
		}
	}
}
