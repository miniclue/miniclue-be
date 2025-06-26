package repository

import (
	"app/internal/model"
	"context"
	"database/sql"
	"fmt"
)

// ExplanationRepository defines explanation-related DB operations
type ExplanationRepository interface {
	GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error)
	CreateExplanationByLectureID(ctx context.Context, lectureID string, slideNumber int, content string) (*model.Explanation, error)
}

// explanationRepository is the DB implementation of ExplanationRepository
type explanationRepository struct {
	db *sql.DB
}

// NewExplanationRepository creates a new ExplanationRepository
func NewExplanationRepository(db *sql.DB) ExplanationRepository {
	return &explanationRepository{db: db}
}

// GetExplanationsByLectureID retrieves explanation records for a given lecture with pagination
func (r *explanationRepository) GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error) {
	query := `SELECT id, lecture_id, slide_number, content, created_at, updated_at FROM explanations WHERE lecture_id = $1 ORDER BY slide_number LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, lectureID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query explanations: %w", err)
	}
	defer rows.Close()

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

// CreateExplanationByLectureID creates an explanation record for the given lecture and slide
func (r *explanationRepository) CreateExplanationByLectureID(ctx context.Context, lectureID string, slideNumber int, content string) (*model.Explanation, error) {
	query := `INSERT INTO explanations (lecture_id, slide_number, content) VALUES ($1, $2, $3) RETURNING id, lecture_id, slide_number, content, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query, lectureID, slideNumber, content)
	var e model.Explanation
	if err := row.Scan(&e.ID, &e.LectureID, &e.SlideNumber, &e.Content, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create explanation: %w", err)
	}
	return &e, nil
}
