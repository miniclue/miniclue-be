package dto

import "time"

type UserCreateDTO struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type UserResponseDTO struct {
	UserID           string                     `json:"user_id"`
	Name             string                     `json:"name"`
	Email            string                     `json:"email"`
	AvatarURL        string                     `json:"avatar_url"`
	APIKeysProvided  map[string]bool            `json:"api_keys_provided"`
	ModelPreferences map[string]map[string]bool `json:"model_preferences"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
}

type APIKeyRequestDTO struct {
	Provider string `json:"provider" validate:"required,oneof=openai gemini anthropic xai deepseek"`
	APIKey   string `json:"api_key" validate:"required"`
}

type APIKeyResponseDTO struct {
	Provider       string `json:"provider"`
	HasProvidedKey bool   `json:"has_provided_key"`
}

type ModelPreferenceRequestDTO struct {
	Provider string `json:"provider" validate:"required,oneof=openai gemini anthropic xai deepseek"`
	Model    string `json:"model" validate:"required"`
	Enabled  bool   `json:"enabled"`
}

// ModelToggleDTO represents a single model and whether it is enabled.
type ModelToggleDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// ProviderModelsDTO contains the models for a provider.
type ProviderModelsDTO struct {
	Provider string           `json:"provider"`
	Models   []ModelToggleDTO `json:"models"`
}

// ModelsResponseDTO wraps providers and their available models.
type ModelsResponseDTO struct {
	Providers []ProviderModelsDTO `json:"providers"`
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
