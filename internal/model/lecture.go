package model

import "time"

// Lecture represents the metadata for an uploaded PDF lecture.
type Lecture struct {
	ID          string    `db:"id" json:"id"`           // UUID or unique string
	UserID      string    `db:"user_id" json:"user_id"` // Supabase Auth user UUID
	CourseID    string    `db:"course_id" json:"course_id"`
	Title       string    `db:"title" json:"title"`
	StoragePath string    `db:"storage_path" json:"storage_path"`
	Status      string    `db:"status" json:"status"` // e.g., "uploaded", "parsed", "explained"
	TotalSlides int       `db:"total_slides" json:"total_slides"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	AccessedAt  time.Time `db:"accessed_at" json:"accessed_at"`
}

// Note represents a user's saved note or snippet from chat.
type Note struct {
	ID        string    `db:"id" json:"id"`                 // UUID
	UserID    string    `db:"user_id" json:"user_id"`       // foreign key to Supabase Auth user
	LectureID string    `db:"lecture_id" json:"lecture_id"` // foreign key
	Content   string    `db:"content" json:"content"`       // rich text/HTML or markdown
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
