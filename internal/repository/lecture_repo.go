package repository

import (
	"context"
	"database/sql"
	"fmt"

	"app/internal/model"

	"github.com/rs/zerolog"
)

type LectureRepository interface {
	GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error)
}

type lectureRepository struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewLectureRepository(db *sql.DB, logger zerolog.Logger) LectureRepository {
	return &lectureRepository{db: db, logger: logger}
}

func (r *lectureRepository) GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE user_id = $1
		ORDER BY accessed_at DESC
		LIMIT %d OFFSET %d
	`, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent lectures: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	var lectures []model.Lecture
	for rows.Next() {
		var lecture model.Lecture
		if err := rows.Scan(
			&lecture.ID,
			&lecture.UserID,
			&lecture.CourseID,
			&lecture.Title,
			&lecture.StoragePath,
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
	query := fmt.Sprintf(`
		SELECT id, user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE course_id = $1
		ORDER BY accessed_at DESC
		LIMIT %d OFFSET %d
	`, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query lectures by course: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error().Err(err).Msg("Failed to close rows")
		}
	}()

	var lectures []model.Lecture
	for rows.Next() {
		var lecture model.Lecture
		if err := rows.Scan(
			&lecture.ID,
			&lecture.UserID,
			&lecture.CourseID,
			&lecture.Title,
			&lecture.StoragePath,
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
		SELECT id, user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
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
		&lecture.StoragePath,
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
		SET title = $1, accessed_at = $2, storage_path = $3, status = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
	`
	return r.db.QueryRowContext(ctx, query,
		l.Title, l.AccessedAt, l.StoragePath, l.Status, l.ID,
	).Scan(
		&l.UserID,
		&l.CourseID,
		&l.Title,
		&l.StoragePath,
		&l.Status,
		&l.CreatedAt,
		&l.UpdatedAt,
		&l.AccessedAt,
	)
}

func (r *lectureRepository) CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error) {
	query := `INSERT INTO lectures (course_id, user_id, title, status, storage_path) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at, accessed_at`
	err := r.db.QueryRowContext(ctx, query, lecture.CourseID, lecture.UserID, lecture.Title, lecture.Status, lecture.StoragePath).Scan(&lecture.ID, &lecture.CreatedAt, &lecture.UpdatedAt, &lecture.AccessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create lecture: %w", err)
	}
	return lecture, nil
}
