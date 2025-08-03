package dto

// LectureUploadURLRequestDTO is the request body for getting upload URLs.
type LectureUploadURLRequestDTO struct {
	CourseID  string   `json:"course_id" validate:"required"`
	Filenames []string `json:"filenames" validate:"required,min=1,max=10"`
}

// LectureUploadURLResponseDTO is the response for a successful upload URL request.
type LectureUploadURLResponseDTO struct {
	LectureID string `json:"lecture_id"`
	UploadURL string `json:"upload_url"`
}

// LectureBatchUploadURLResponseDTO is the response for a successful batch upload URL request.
type LectureBatchUploadURLResponseDTO struct {
	Uploads []LectureUploadURLResponseDTO `json:"uploads"`
}
