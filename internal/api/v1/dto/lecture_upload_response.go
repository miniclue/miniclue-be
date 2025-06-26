package dto

// LectureUploadResponseDTO is returned when a PDF is successfully uploaded to create a lecture
// swagger:model LectureUploadResponseDTO
// example:
//
//	lecture_id: "20000000-0000-0000-0000-000000000001"
//	status: "uploaded"
type LectureUploadResponseDTO struct {
	// The ID of the newly created lecture
	LectureID string `json:"lecture_id"`
	// The status of the lecture
	Status string `json:"status"`
}
