package repository

import (
	"app/internal/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog"
)

// ExplanationRepository defines explanation-related DB operations
type ExplanationRepository interface {
	GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error)
}

// explanationRepository is the DB implementation of ExplanationRepository
type explanationRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewExplanationRepository creates a new ExplanationRepository
func NewExplanationRepository(db *sql.DB, logger zerolog.Logger) ExplanationRepository {
	return &explanationRepository{db: db, logger: logger}
}

// GetExplanationsByLectureID retrieves explanation records for a given lecture with pagination
func (r *explanationRepository) GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error) {
	baseQuery := `SELECT id, lecture_id, slide_number, content, created_at, updated_at FROM explanations WHERE lecture_id = $1 ORDER BY slide_number`
	var (
		rows *sql.Rows
		err  error
	)
	if limit > 0 {
		query := baseQuery + " LIMIT $2 OFFSET $3"
		rows, err = r.db.QueryContext(ctx, query, lectureID, limit, offset)
	} else if offset > 0 {
		query := baseQuery + " OFFSET $2"
		rows, err = r.db.QueryContext(ctx, query, lectureID, offset)
	} else {
		rows, err = r.db.QueryContext(ctx, baseQuery, lectureID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query explanations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	explanations := []model.Explanation{}
	for rows.Next() {
		var e model.Explanation
		if err := rows.Scan(&e.ID, &e.LectureID, &e.SlideNumber, &e.Content, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan explanation: %w", err)
		}
		explanations = append(explanations, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return explanations, nil
}
