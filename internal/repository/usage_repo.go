package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UsageRepository tracks user actions for usage-based limits.
type UsageRepository interface {
	// RecordUploadEvent records a new upload event without checking quota limits.
	RecordUploadEvent(ctx context.Context, userID string, lectureID string) error
	// CountUploadEventsInTimeRange counts lecture uploads in the given period.
	CountUploadEventsInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error)
}

type usageRepo struct {
	pool *pgxpool.Pool
}

// NewUsageRepo creates a new UsageRepository.
func NewUsageRepo(pool *pgxpool.Pool) UsageRepository {
	return &usageRepo{pool: pool}
}

// CountUploadEventsInTimeRange counts lecture uploads in the given period.
func (r *usageRepo) CountUploadEventsInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error) {
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

// RecordUploadEvent records a new upload event without checking quota limits.
func (r *usageRepo) RecordUploadEvent(ctx context.Context, userID string, lectureID string) error {
	const insertQ = `INSERT INTO usage_events (user_id, event_type, lecture_id) VALUES ($1, 'lecture_upload', $2)`
	if _, err := r.pool.Exec(ctx, insertQ, userID, lectureID); err != nil {
		return fmt.Errorf("recording upload event for user %s: %w", userID, err)
	}
	return nil
}
