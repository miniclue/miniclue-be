package dto

// LectureUploadCompleteRequestDTO is the request body for completing an upload.
type LectureUploadCompleteRequestDTO struct {
	// No fields needed - the lecture ID is in the URL path
}

// LectureUploadCompleteResponseDTO is the response for a successful upload completion.
type LectureUploadCompleteResponseDTO struct {
	LectureID string `json:"lecture_id"`
	CourseID  string `json:"course_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}
