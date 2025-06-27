package dto

import "time"

type LectureResponseDTO struct {
	LectureID  string    `json:"lecture_id"`
	CourseID   string    `json:"course_id"`
	Title      string    `json:"title"`
	PdfURL     string    `json:"pdf_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AccessedAt time.Time `json:"accessed_at"`
}

type LectureUpdateDTO struct {
	Title      *string    `json:"title,omitempty"`
	AccessedAt *time.Time `json:"accessed_at,omitempty"`
}
