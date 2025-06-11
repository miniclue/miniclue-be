package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// LectureService defines lecture-related operations
// GetLecturesByCourseID retrieves lectures for a given course with pagination
type LectureService interface {
	GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error)
	GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error)
}

// lectureService is the implementation of LectureService
type lectureService struct {
	repo repository.LectureRepository
}

// NewLectureService creates a new LectureService
func NewLectureService(repo repository.LectureRepository) LectureService {
	return &lectureService{repo: repo}
}

// GetLecturesByCourseID retrieves lectures for a given course with pagination
func (s *lectureService) GetLecturesByCourseID(ctx context.Context, courseID string, limit, offset int) ([]model.Lecture, error) {
	return s.repo.GetLecturesByCourseID(ctx, courseID, limit, offset)
}

// GetLectureByID retrieves a lecture by ID
func (s *lectureService) GetLectureByID(ctx context.Context, lectureID string) (*model.Lecture, error) {
	return s.repo.GetLectureByID(ctx, lectureID)
}
