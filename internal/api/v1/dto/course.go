package dto

// CourseResponseDTO is returned in API responses for course data
type CourseResponseDTO struct {
	CourseID    string `json:"course_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
