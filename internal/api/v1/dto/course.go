package dto

import "time"

// CourseCreateDTO is used for incoming course creation requests
type CourseCreateDTO struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

// CourseResponseDTO is returned in API responses for courses
type CourseResponseDTO struct {
	CourseID    string    `json:"course_id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CourseUpdateDTO is used for incoming course update requests
type CourseUpdateDTO struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}
