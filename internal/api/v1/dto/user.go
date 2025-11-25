package dto

import "time"

type UserCreateDTO struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type UserResponseDTO struct {
	UserID          string          `json:"user_id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	AvatarURL       string          `json:"avatar_url"`
	APIKeysProvided map[string]bool `json:"api_keys_provided"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type APIKeyRequestDTO struct {
	Provider string `json:"provider" validate:"required,oneof=openai gemini"`
	APIKey   string `json:"api_key" validate:"required"`
}

type APIKeyResponseDTO struct {
	Provider       string `json:"provider"`
	HasProvidedKey bool   `json:"has_provided_key"`
}

type UserCourseResponseDTO struct {
	CourseID    string    `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserRecentLectureResponseDTO struct {
	LectureID string `json:"lecture_id"`
	Title     string `json:"title"`
	CourseID  string `json:"course_id"`
}

type UserRecentLecturesResponseDTO struct {
	Lectures   []UserRecentLectureResponseDTO `json:"lectures"`
	TotalCount int                            `json:"total_count"`
}
