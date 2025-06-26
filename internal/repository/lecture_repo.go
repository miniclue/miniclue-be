package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"
)

type LectureRepository interface {
	GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error)
	EnqueueIngestionJob(ctx context.Context, lectureID string, storagePath string) error
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

func (r *lectureRepository) GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error) {
	query := `
		SELECT id, user_id, course_id, title, pdf_url, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE course_id = $1
		ORDER BY accessed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, courseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query lectures by course: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return lectures, nil
}

func (r *lectureRepository) GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error) {
	query := `
		SELECT id, user_id, course_id, title, pdf_url, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, lectureID)
	var lecture model.Lecture
	if err := row.Scan(
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan lecture row: %w", err)
	}
	return &lecture, nil
}

func (r *lectureRepository) DeleteLecture(ctx context.Context, lectureID string) error {
	query := `DELETE FROM lectures WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, lectureID); err != nil {
		return fmt.Errorf("failed to delete lecture: %w", err)
	}
	return nil
}

func (r *lectureRepository) UpdateLecture(ctx context.Context, l *model.Lecture) error {
	query := `
		UPDATE lectures
		SET title = $1, accessed_at = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING user_id, course_id, title, pdf_url, status, created_at, updated_at, accessed_at
	`
	return r.db.QueryRowContext(ctx, query,
		l.Title, l.AccessedAt, l.ID,
	).Scan(
		&l.UserID,
		&l.CourseID,
		&l.Title,
		&l.PDFURL,
		&l.Status,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.AccessedAt,
	)
}

func (r *lectureRepository) CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error) {
	query := `INSERT INTO lectures (course_id, user_id, title, status, pdf_url) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at, accessed_at`
	err := r.db.QueryRowContext(ctx, query, lecture.CourseID, lecture.UserID, lecture.Title, lecture.Status, lecture.PDFURL).Scan(&lecture.ID, &lecture.CreatedAt, &lecture.UpdatedAt, &lecture.AccessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create lecture: %w", err)
	}
	return lecture, nil
}

func (r *lectureRepository) EnqueueIngestionJob(ctx context.Context, lectureID string, storagePath string) error {
	query := `SELECT pgmq.send('ingestion_queue', $1::jsonb, 0)`
	job := fmt.Sprintf(`{"lecture_id": "%s", "storage_path": "%s"}`, lectureID, storagePath)
	_, err := r.db.ExecContext(ctx, query, job)
	if err != nil {
		return fmt.Errorf("failed to enqueue ingestion job: %w", err)
	}
	return nil
}
