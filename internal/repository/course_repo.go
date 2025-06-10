package repository

import (
	"context"
	"database/sql"

	"app/internal/model"
)

// CourseRepository defines the interface for interacting with course data
type CourseRepository interface {
	GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error)
	CreateCourse(ctx context.Context, c *model.Course) error
	// GetCourseByID retrieves a course by its ID
	GetCourseByID(ctx context.Context, courseID string) (*model.Course, error)
	// UpdateCourse updates an existing course
	UpdateCourse(ctx context.Context, c *model.Course) error
}

type courseRepo struct {
	db *sql.DB
}

// NewCourseRepo creates a new CourseRepository
func NewCourseRepo(db *sql.DB) CourseRepository {
	return &courseRepo{db: db}
}

// GetCoursesByUserID retrieves all courses associated with a given user ID
func (r *courseRepo) GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error) {
	var courses []model.Course
	query := `
		SELECT id, title, description
		FROM courses
		WHERE user_id = $1
		ORDER BY title ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var course model.Course
		if err := rows.Scan(
			&course.CourseID,
			&course.Title,
			&course.Description,
		); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}

	if err = rows.Err(); err != nil {
		return nil, err
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
	return r.db.QueryRowContext(ctx, query, c.UserID, c.Title, c.Description, c.IsDefault).
		Scan(&c.CourseID, &c.UserID, &c.Title, &c.Description, &c.IsDefault, &c.CreatedAt, &c.UpdatedAt)
}

// GetCourseByID retrieves a course by its ID
func (r *courseRepo) GetCourseByID(ctx context.Context, courseID string) (*model.Course, error) {
	query := `
		SELECT id, user_id, title, description, is_default, created_at, updated_at
		FROM courses
		WHERE id = $1
	`
	var c model.Course
	err := r.db.QueryRowContext(ctx, query, courseID).Scan(
		&c.CourseID,
		&c.UserID,
		&c.Title,
		&c.Description,
		&c.IsDefault,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
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
	return r.db.QueryRowContext(ctx, query, c.Title, c.Description, c.IsDefault, c.CourseID).
		Scan(&c.UserID, &c.Title, &c.Description, &c.IsDefault, &c.CreatedAt, &c.UpdatedAt)
}
