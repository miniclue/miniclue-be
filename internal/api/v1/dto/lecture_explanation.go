package dto

import "time"

// LectureExplanationResponseDTO represents explanation for a specific slide
// @Summary Lecture explanation
// @Tags lectures
// @Produce json
// @Success 200 {array} dto.LectureExplanationResponseDTO
// @Router /lectures/{lectureId}/explanations [get]
type LectureExplanationResponseDTO struct {
    ID          string    `json:"id"`
    LectureID   string    `json:"lecture_id"`
    SlideNumber int       `json:"slide_number"`
    Content     string    `json:"content"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
