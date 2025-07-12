package model

import "time"

// DeadLetterMessage represents a message from the dead-letter queue persisted in the database.
type DeadLetterMessage struct {
	ID               string    `db:"id"`
	SubscriptionName string    `db:"subscription_name"`
	MessageID        string    `db:"message_id"`
	Payload          string    `db:"payload"`    // Should be a JSON string
	Attributes       *string   `db:"attributes"` // Can be null, should be a JSON string
	Status           string    `db:"status"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
