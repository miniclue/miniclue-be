package dto

// LectureUploadResponseDTO represents the data returned for each uploaded file.
type LectureUploadResponseDTO struct {
	Filename  string `json:"filename"`
	LectureID string `json:"lecture_id"`
	Status    string `json:"status"`
}
