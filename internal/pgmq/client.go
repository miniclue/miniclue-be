package pgmq

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// Client wraps a Postgres DB for pgmq queue operations.
type Client struct {
	db *sql.DB
}

// New returns a new PGMQ client backed by the given DB connection.
func New(db *sql.DB) *Client {
	return &Client{db: db}
}

// Message represents a single pgmq message.
type Message struct {
	ID   int64  // message identifier
	Data []byte // raw JSON payload
}

// Send pushes a JSON payload into the given queue.
func (c *Client) Send(ctx context.Context, queue string, payload []byte) error {
	query := "SELECT pgmq.send($1, $2::jsonb, 0)"
	if _, err := c.db.ExecContext(ctx, query, queue, string(payload)); err != nil {
		return fmt.Errorf("pgmq send failed: %w", err)
	}
	return nil
}

// ReadWithPoll reads up to maxMessages from the queue, blocking up to timeoutSec seconds.
func (c *Client) ReadWithPoll(ctx context.Context, queue string, timeoutSec, maxMessages int) ([]*Message, error) {
	query := "SELECT msg_id, message FROM pgmq.read_with_poll($1, $2, $3)"
	rows, err := c.db.QueryContext(ctx, query, queue, timeoutSec, maxMessages)
	if err != nil {
		return nil, fmt.Errorf("pgmq read_with_poll failed: %w", err)
	}
	defer rows.Close()

	var msgs []*Message
	for rows.Next() {
		var id int64
		var data []byte
		if err := rows.Scan(&id, &data); err != nil {
			return nil, fmt.Errorf("pgmq read scan failed: %w", err)
		}
		msgs = append(msgs, &Message{ID: id, Data: data})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgmq read rows error: %w", err)
	}
	return msgs, nil
}

// Delete removes messages by their IDs from the specified queue.
func (c *Client) Delete(ctx context.Context, queue string, msgIDs []int64) error {
	query := "SELECT pgmq.delete($1::text, $2::bigint[])"
	if _, err := c.db.ExecContext(ctx, query, queue, pq.Array(msgIDs)); err != nil {
		return fmt.Errorf("pgmq delete failed: %w", err)
	}
	return nil
}

// Exec runs a SQL query with the given arguments, returning an error if it fails.
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) error {
	if _, err := c.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("db exec failed: %w", err)
	}
	return nil
}

// QueryRow runs a query expected to return at most one row.
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}
