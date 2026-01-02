package repository

import (
	"context"
	"fmt"

	"app/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DLQRepository interface {
	Create(ctx context.Context, message *model.DeadLetterMessage) error
}

type dlqRepository struct {
	pool *pgxpool.Pool
}

func NewDLQRepository(pool *pgxpool.Pool) DLQRepository {
	return &dlqRepository{pool: pool}
}

func (r *dlqRepository) Create(ctx context.Context, message *model.DeadLetterMessage) error {
	query := `
        INSERT INTO dead_letter_messages (subscription_name, message_id, payload, attributes, status)
        VALUES ($1, $2, $3::jsonb, $4::jsonb, $5)
    `
	_, err := r.pool.Exec(
		ctx,
		query,
		message.SubscriptionName,
		message.MessageID,
		string(message.Payload),
		string(message.Attributes),
		message.Status,
	)
	if err != nil {
		return fmt.Errorf("creating dead letter message for subscription %s: %w", message.SubscriptionName, err)
	}
	return nil
}
