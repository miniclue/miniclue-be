package dto

import "time"

// LectureNoteResponseDTO represents a note for a specific lecture
// @Summary Lecture note
// @Tags lectures
// @Produce json
// @Success 200 {array} dto.LectureNoteResponseDTO
// @Router /lectures/{lectureId}/notes [get]
type LectureNoteResponseDTO struct {
	ID        string    `json:"id"`
	LectureID string    `json:"lecture_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
