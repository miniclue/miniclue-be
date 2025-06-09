package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"
)

type LectureRepository interface {
	GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
}

type lectureRepository struct {
	db *sql.DB
}

func NewLectureRepository(db *sql.DB) LectureRepository {
	return &lectureRepository{db: db}
}

func (r *lectureRepository) GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error) {
	query := `
		SELECT id, user_id, course_id, title, pdf_url, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE user_id = $1
		ORDER BY accessed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent lectures: %w", err)
	}
	defer rows.Close()

	var lectures []model.Lecture
	for rows.Next() {
		var lecture model.Lecture
		if err := rows.Scan(
			&lecture.ID,
			&lecture.UserID,
			&lecture.CourseID,
			&lecture.Title,
			&lecture.PDFURL,
			&lecture.Status,
			&lecture.CreatedAt,
			&lecture.UpdatedAt,
			&lecture.AccessedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lecture row: %w", err)
		}
		lectures = append(lectures, lecture)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return lectures, nil
}