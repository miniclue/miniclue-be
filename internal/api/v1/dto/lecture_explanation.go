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

// LectureExplanationCreateDTO represents payload to create an explanation for a specific lecture
// @Summary Create lecture explanation
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param explanation body LectureExplanationCreateDTO true "Explanation create data"
// @Success 201 {object} dto.LectureExplanationResponseDTO
// @Router /lectures/{lectureId}/explanations [post]
type LectureExplanationCreateDTO struct {
	SlideNumber int    `json:"slide_number" validate:"required"`
	Content     string `json:"content" validate:"required"`
}
