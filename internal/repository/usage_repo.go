package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UsageRepository tracks user actions for usage-based limits.
type UsageRepository interface {
	RecordUploadEvent(ctx context.Context, userID string) error
	CountUploadEvents(ctx context.Context, userID string, start, end time.Time) (int, error)
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
