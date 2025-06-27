package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// ExplanationService defines explanation-related operations
// GetExplanationsByLectureID retrieves explanations for a given lecture with pagination, returns empty slice if none
type ExplanationService interface {
	GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error)
}

// explanationService is the implementation of ExplanationService
type explanationService struct {
	repo repository.ExplanationRepository
}

// NewExplanationService creates a new ExplanationService
func NewExplanationService(repo repository.ExplanationRepository) ExplanationService {
	return &explanationService{repo: repo}
}

// GetExplanationsByLectureID retrieves explanations for a given lecture with pagination
func (s *explanationService) GetExplanationsByLectureID(ctx context.Context, lectureID string, limit, offset int) ([]model.Explanation, error) {
	return s.repo.GetExplanationsByLectureID(ctx, lectureID, limit, offset)
}
