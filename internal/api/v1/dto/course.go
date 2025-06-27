package dto

import "time"

type CourseCreateDTO struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

type CourseResponseDTO struct {
	CourseID    string    `json:"course_id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CourseUpdateDTO struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
}

type CourseLectureResponseDTO struct {
	LectureID string `json:"lecture_id"`
	Title     string `json:"title"`
}
