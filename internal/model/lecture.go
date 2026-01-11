package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Lecture represents the metadata for an uploaded PDF lecture.
type Lecture struct {
	ID                    string                `db:"id" json:"id"`           // UUID or unique string
	UserID                string                `db:"user_id" json:"user_id"` // Supabase Auth user UUID
	CourseID              string                `db:"course_id" json:"course_id"`
	Title                 string                `db:"title" json:"title"`
	StoragePath           string                `db:"storage_path" json:"storage_path"`
	Status                string                `db:"status" json:"status"` // e.g., "uploaded", "parsed", "explained"
	EmbeddingErrorDetails EmbeddingErrorDetails `db:"embedding_error_details" json:"embedding_error_details"`
	TotalSlides           int                   `db:"total_slides" json:"total_slides"`
	CreatedAt             time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time             `db:"updated_at" json:"updated_at"`
	AccessedAt            time.Time             `db:"accessed_at" json:"accessed_at"`
}

// EmbeddingErrorDetails is a map for storing error details (JSONB)
type EmbeddingErrorDetails map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (e EmbeddingErrorDetails) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

// Scan implements the sql.Scanner interface for JSONB
func (e *EmbeddingErrorDetails) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*e = nil
		return fmt.Errorf("cannot scan %T into EmbeddingErrorDetails", value)
	}

	if len(bytes) == 0 {
		*e = nil
		return nil
	}

	return json.Unmarshal(bytes, e)
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
