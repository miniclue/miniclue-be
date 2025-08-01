package dto

import "time"

type UserCreateDTO struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type UserResponseDTO struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
}

type UserRecentLecturesResponseDTO struct {
	Lectures   []UserRecentLectureResponseDTO `json:"lectures"`
	TotalCount int                            `json:"total_count"`
}

type UserUsageResponseDTO struct {
	UserID             string    `json:"user_id"`
	CurrentUsage       int       `json:"current_usage"`
	MaxUploads         int       `json:"max_uploads"`
	MaxSizeMB          int       `json:"max_size_mb"`
	PlanID             string    `json:"plan_id"`
	BillingPeriodStart time.Time `json:"billing_period_start"`
	BillingPeriodEnd   time.Time `json:"billing_period_end"`
	PlanName           string    `json:"plan_name"`
}
