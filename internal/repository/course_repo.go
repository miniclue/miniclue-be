package repository

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseRepository defines the interface for interacting with course data
type CourseRepository interface {
	GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error)
	CreateCourse(ctx context.Context, c *model.Course) error
	// GetCourseByID retrieves a course by its ID
	GetCourseByID(ctx context.Context, courseID string) (*model.Course, error)
	GetDefaultCourseByUserID(ctx context.Context, userID string) (*model.Course, error)
	// UpdateCourse updates an existing course
	UpdateCourse(ctx context.Context, c *model.Course) error
	// DeleteCourse deletes a course by its ID
	DeleteCourse(ctx context.Context, courseID string) error
}

type courseRepo struct {
	pool *pgxpool.Pool
}

// NewCourseRepo creates a new CourseRepository
func NewCourseRepo(pool *pgxpool.Pool) CourseRepository {
	return &courseRepo{pool: pool}
}

// GetCoursesByUserID retrieves all courses associated with a given user ID
func (r *courseRepo) GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error) {
	var courses []model.Course
	query := `
		SELECT id, title, description, is_default, updated_at
		FROM courses
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying courses for user %s: %w", userID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var course model.Course
		if err := rows.Scan(
			&course.CourseID,
			&course.Title,
			&course.Description,
			&course.IsDefault,
			&course.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning course row: %w", err)
		}
		courses = append(courses, course)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating course rows: %w", err)
	}

	// If no courses found, return an empty slice, not nil
	if len(courses) == 0 {
		return []model.Course{}, nil
	}

	return courses, nil
}

// CreateCourse inserts a new course and returns the created record
func (r *courseRepo) CreateCourse(ctx context.Context, c *model.Course) error {
	query := `
		INSERT INTO courses (user_id, title, description, is_default)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, description, is_default, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, c.UserID, c.Title, c.Description, c.IsDefault).
		Scan(&c.CourseID, &c.UserID, &c.Title, &c.Description, &c.IsDefault, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating course: %w", err)
	}
	return nil
}

// GetCourseByID retrieves a course by its ID
func (r *courseRepo) GetCourseByID(ctx context.Context, courseID string) (*model.Course, error) {
	query := `
		SELECT id, user_id, title, description, is_default, created_at, updated_at
		FROM courses
		WHERE id = $1
	`
	var c model.Course
	err := r.pool.QueryRow(ctx, query, courseID).Scan(
		&c.CourseID,
		&c.UserID,
		&c.Title,
		&c.Description,
		&c.IsDefault,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting course by id %s: %w", courseID, err)
	}
	return &c, nil
}

func (r *courseRepo) GetDefaultCourseByUserID(ctx context.Context, userID string) (*model.Course, error) {
	query := `
		SELECT id, user_id, title, description, is_default, created_at, updated_at
		FROM courses
		WHERE user_id = $1 AND is_default = TRUE
		LIMIT 1
	`
	var c model.Course
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&c.CourseID,
		&c.UserID,
		&c.Title,
		&c.Description,
		&c.IsDefault,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting default course by user id %s: %w", userID, err)
	}
	return &c, nil
}

// UpdateCourse updates an existing course record and returns updated timestamps
func (r *courseRepo) UpdateCourse(ctx context.Context, c *model.Course) error {
	query := `
		UPDATE courses
		SET title = $1, description = $2, is_default = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING user_id, title, description, is_default, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, c.Title, c.Description, c.IsDefault, c.CourseID).
		Scan(&c.UserID, &c.Title, &c.Description, &c.IsDefault, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating course %s: %w", c.CourseID, err)
	}
	return nil
}

// DeleteCourse deletes a course and cascades to related records via DB ON DELETE CASCADE
func (r *courseRepo) DeleteCourse(ctx context.Context, courseID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM courses WHERE id = $1`, courseID)
	if err != nil {
		return fmt.Errorf("deleting course %s: %w", courseID, err)
	}
	return nil
}
