package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUploadLimitExceeded is returned when a user has reached their upload limit.
var ErrUploadLimitExceeded = errors.New("upload_limit_exceeded")

// UsageRepository tracks user actions for usage-based limits.
type UsageRepository interface {
	RecordUploadEvent(ctx context.Context, userID string) error
	CountUploadEvents(ctx context.Context, userID string, start, end time.Time) (int, error)
	// CheckAndRecordUpload atomically checks the user's upload count for the period and records a new upload. Returns ErrUploadLimitExceeded if the limit is reached.
	CheckAndRecordUpload(ctx context.Context, userID string, start, end time.Time, maxUploads int) error
}

type usageRepo struct {
	pool *pgxpool.Pool
}

// NewUsageRepo creates a new UsageRepository.
func NewUsageRepo(pool *pgxpool.Pool) UsageRepository {
	return &usageRepo{pool: pool}
}

// RecordUploadEvent logs a lecture upload event for a user.
func (r *usageRepo) RecordUploadEvent(ctx context.Context, userID string) error {
	const q = `INSERT INTO usage_events (user_id, event_type) VALUES ($1, 'lecture_upload')`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("recording upload event for user %s: %w", userID, err)
	}
	return nil
}

// CountUploadEvents counts lecture uploads in the given period.
func (r *usageRepo) CountUploadEvents(ctx context.Context, userID string, start, end time.Time) (int, error) {
	var count int
	const q = `
        SELECT COUNT(*)
        FROM usage_events
        WHERE user_id = $1
          AND event_type = 'lecture_upload'
          AND created_at >= $2
          AND created_at < $3
    `
	if err := r.pool.QueryRow(ctx, q, userID, start, end).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting upload events for user %s: %w", userID, err)
	}
	return count, nil
}

// CheckAndRecordUpload atomically checks the user's upload count for the period and records a new upload event.
func (r *usageRepo) CheckAndRecordUpload(ctx context.Context, userID string, start, end time.Time, maxUploads int) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("starting transaction for upload check: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	var count int
	const countQ = `
		SELECT COUNT(*)
		FROM usage_events
		WHERE user_id = $1
		  AND event_type = 'lecture_upload'
		  AND created_at >= $2
		  AND created_at < $3
	`
	if err := tx.QueryRow(ctx, countQ, userID, start, end).Scan(&count); err != nil {
		return fmt.Errorf("counting uploads for user %s: %w", userID, err)
	}
	if maxUploads > 0 && count >= maxUploads {
		return ErrUploadLimitExceeded
	}
	const insertQ = `INSERT INTO usage_events (user_id, event_type) VALUES ($1, 'lecture_upload')`
	if _, err := tx.Exec(ctx, insertQ, userID); err != nil {
		return fmt.Errorf("recording upload event for user %s: %w", userID, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing upload event for user %s: %w", userID, err)
	}
	return nil
}
