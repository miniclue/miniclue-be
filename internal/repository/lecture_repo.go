package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LectureRepository interface {
	GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
	DeleteLecture(ctx context.Context, lectureID string) error
	UpdateLecture(ctx context.Context, l *model.Lecture) error
	CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error)
	CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error)
	CountLecturesByUserID(ctx context.Context, userID string) (int, error)
}

type lectureRepository struct {
	pool *pgxpool.Pool
}

func NewLectureRepository(pool *pgxpool.Pool) LectureRepository {
	return &lectureRepository{pool: pool}
}

func (r *lectureRepository) GetLecturesByUserID(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE user_id = $1
		ORDER BY accessed_at DESC
		LIMIT %d OFFSET %d
	`, limit, offset)

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying recent lectures for user %s: %w", userID, err)
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
			&lecture.StoragePath,
			&lecture.Status,
			&lecture.CreatedAt,
			&lecture.UpdatedAt,
			&lecture.AccessedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning lecture row: %w", err)
		}
		lectures = append(lectures, lecture)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recent lecture rows: %w", err)
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

	rows, err := r.pool.Query(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("querying lectures for course %s: %w", courseID, err)
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
			&lecture.StoragePath,
			&lecture.Status,
			&lecture.CreatedAt,
			&lecture.UpdatedAt,
			&lecture.AccessedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning lecture row for course: %w", err)
		}
		lectures = append(lectures, lecture)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating lecture rows for course: %w", err)
	}

	return lectures, nil
}

func (r *lectureRepository) GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error) {
	query := `
		SELECT id, user_id, course_id, title, storage_path, status, created_at, updated_at, accessed_at
		FROM lectures
		WHERE id = $1
	`
	var lecture model.Lecture
	err := r.pool.QueryRow(ctx, query, lectureID).Scan(
		&lecture.ID,
		&lecture.UserID,
		&lecture.CourseID,
		&lecture.Title,
		&lecture.StoragePath,
		&lecture.Status,
		&lecture.CreatedAt,
		&lecture.UpdatedAt,
		&lecture.AccessedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning lecture row for id %s: %w", lectureID, err)
	}
	return &lecture, nil
}

func (r *lectureRepository) DeleteLecture(ctx context.Context, lectureID string) error {
	query := `DELETE FROM lectures WHERE id = $1`
	if _, err := r.pool.Exec(ctx, query, lectureID); err != nil {
		return fmt.Errorf("deleting lecture %s: %w", lectureID, err)
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
	err := r.pool.QueryRow(ctx, query,
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
	if err != nil {
		return fmt.Errorf("updating lecture %s: %w", l.ID, err)
	}
	return nil
}

func (r *lectureRepository) CreateLecture(ctx context.Context, lecture *model.Lecture) (*model.Lecture, error) {
	query := `INSERT INTO lectures (course_id, user_id, title, status, storage_path) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at, accessed_at`
	err := r.pool.QueryRow(ctx, query, lecture.CourseID, lecture.UserID, lecture.Title, lecture.Status, lecture.StoragePath).Scan(&lecture.ID, &lecture.CreatedAt, &lecture.UpdatedAt, &lecture.AccessedAt)
	if err != nil {
		return nil, fmt.Errorf("creating lecture: %w", err)
	}
	return lecture, nil
}

func (r *lectureRepository) CountLecturesByUserInTimeRange(ctx context.Context, userID string, start, end time.Time) (int, error) {
	// Count actual lecture upload events rather than stored lectures
	var count int
	query := `
		SELECT COUNT(*)
		FROM usage_events
		WHERE user_id = $1
			AND event_type = 'lecture_upload'
			AND created_at >= $2
			AND created_at < $3
	`
	err := r.pool.QueryRow(ctx, query, userID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting lectures for user %s: %w", userID, err)
	}
	return count, nil
}

func (r *lectureRepository) CountLecturesByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM lectures WHERE user_id = $1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting lectures for user %s: %w", userID, err)
	}
	return count, nil
}
