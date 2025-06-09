package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// CourseService defines the interface for course-related business logic
type CourseService interface {
	GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error)
}

type courseService struct {
	repo repository.CourseRepository
}

// NewCourseService creates a new CourseService
func NewCourseService(repo repository.CourseRepository) CourseService {
	return &courseService{repo: repo}
}

// GetCoursesByUserID retrieves all courses for a given user
func (s *courseService) GetCoursesByUserID(ctx context.Context, userID string) ([]model.Course, error) {
	// Simply delegate to the repository layer
	courses, err := s.repo.GetCoursesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return courses, nil
}
