package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// SummaryService defines summary-related operations
// GetSummaryByLectureID retrieves a lecture's summary, returns nil,nil if none exists
type SummaryService interface {
	GetSummaryByLectureID(ctx context.Context, lectureID string) (*model.Summary, error)
}

// summaryService is the implementation of SummaryService
type summaryService struct {
	repo repository.SummaryRepository
}

// NewSummaryService creates a new SummaryService
func NewSummaryService(repo repository.SummaryRepository) SummaryService {
	return &summaryService{repo: repo}
}

// GetSummaryByLectureID retrieves a lecture's summary
func (s *summaryService) GetSummaryByLectureID(ctx context.Context, lectureID string) (*model.Summary, error) {
	return s.repo.GetSummaryByLectureID(ctx, lectureID)
}
