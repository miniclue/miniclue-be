package dto

import "time"

type LectureNoteResponseDTO struct {
	ID        string    `json:"id"`
	LectureID string    `json:"lecture_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LectureNoteUpdateDTO struct {
	Content string `json:"content" validate:"required"`
}

type LectureNoteCreateDTO struct {
	Content string `json:"content" validate:"required"`
}
