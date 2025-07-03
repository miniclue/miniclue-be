package dto

// SignedURLResponseDTO represents a JSON response containing a signed URL for downloading a PDF.
type SignedURLResponseDTO struct {
	URL string `json:"url"`
}
