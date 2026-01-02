package dto

import "time"

type LectureResponseDTO struct {
	LectureID   string    `json:"lecture_id"`
	CourseID    string    `json:"course_id"`
	Title       string    `json:"title"`
	StoragePath string    `json:"storage_path"`
	Status      string    `json:"status"`
	TotalSlides int       `json:"total_slides"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AccessedAt  time.Time `json:"accessed_at"`
}

type LectureUpdateDTO struct {
	Title      *string    `json:"title,omitempty"`
	AccessedAt *time.Time `json:"accessed_at,omitempty"`
	CourseID   *string    `json:"course_id,omitempty"`
}
