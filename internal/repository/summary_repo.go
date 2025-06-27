package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"
)

// SummaryRepository defines summary-related DB operations
type SummaryRepository interface {
	GetSummaryByLectureID(ctx context.Context, lectureID string) (*model.Summary, error)
}

// summaryRepository is the DB implementation of SummaryRepository
type summaryRepository struct {
	db *sql.DB
}

// NewSummaryRepository creates a new SummaryRepository
func NewSummaryRepository(db *sql.DB) SummaryRepository {
	return &summaryRepository{db: db}
}

// GetSummaryByLectureID retrieves the summary record for a given lecture
func (r *summaryRepository) GetSummaryByLectureID(ctx context.Context, lectureID string) (*model.Summary, error) {
	query := `SELECT lecture_id, content, created_at, updated_at FROM summaries WHERE lecture_id = $1`
	row := r.db.QueryRowContext(ctx, query, lectureID)
	var s model.Summary
	if err := row.Scan(&s.LectureID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan summary: %w", err)
	}
	return &s, nil
}
