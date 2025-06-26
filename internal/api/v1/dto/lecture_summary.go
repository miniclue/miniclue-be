package dto

// LectureSummaryResponseDTO is returned for a lecture's summary
// @Summary Lecture summary
// @Tags lectures
// @Produce json
// @Success 200 {object} dto.LectureSummaryResponseDTO
// @Router /lectures/{lectureId}/summary [get]
type LectureSummaryResponseDTO struct {
	LectureID string `json:"lecture_id"`
	Content   string `json:"content"`
}

// LectureSummaryCreateDTO represents payload to create a summary for a specific lecture
// @Summary Create lecture summary
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param summary body LectureSummaryCreateDTO true "Summary create data"
// @Success 201 {object} dto.LectureSummaryResponseDTO
// @Router /lectures/{lectureId}/summary [post]
type LectureSummaryCreateDTO struct {
	Content string `json:"content" validate:"required"`
}
