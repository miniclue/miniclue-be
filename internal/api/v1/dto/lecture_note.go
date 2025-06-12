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

// LectureNoteUpdateDTO represents payload to update a note's content
// @Summary Lecture note update
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param noteId path string true "Note ID"
// @Param note body LectureNoteUpdateDTO true "Note update data"
type LectureNoteUpdateDTO struct {
	Content string `json:"content" validate:"required"`
}

// LectureNoteCreateDTO represents payload to create a note for a specific lecture
// @Summary Create lecture note
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param note body LectureNoteCreateDTO true "Note create data"
// @Success 201 {object} dto.LectureNoteResponseDTO
// @Router /lectures/{lectureId}/notes [post]
type LectureNoteCreateDTO struct {
	Content string `json:"content" validate:"required"`
}
