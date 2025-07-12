package repository

import (
	"context"
	"database/sql"

	"app/internal/model"
)

type DLQRepository interface {
	Create(ctx context.Context, message *model.DeadLetterMessage) error
}

type dlqRepository struct {
	db *sql.DB
}

func NewDLQRepository(db *sql.DB) DLQRepository {
	return &dlqRepository{db: db}
}

func (r *dlqRepository) Create(ctx context.Context, message *model.DeadLetterMessage) error {
	query := `
        INSERT INTO dead_letter_messages (subscription_name, message_id, payload, attributes, status)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(
		ctx,
		query,
		message.SubscriptionName,
		message.MessageID,
		message.Payload,
		message.Attributes,
		message.Status,
	)
	return err
}
