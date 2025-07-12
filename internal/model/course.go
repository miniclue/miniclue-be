package model

import "time"

// Course represents a course in the system
type Course struct {
	CourseID    string    `db:"id" json:"course_id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	IsDefault   bool      `db:"is_default" json:"is_default"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
