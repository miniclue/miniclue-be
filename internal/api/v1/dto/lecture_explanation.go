package dto

import "time"

type LectureExplanationResponseDTO struct {
	ID          string    `json:"id"`
	LectureID   string    `json:"lecture_id"`
	SlideNumber int       `json:"slide_number"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
