package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// CourseService defines the interface for course operations
type CourseService interface {
	CreateCourse(ctx context.Context, c *model.Course) (*model.Course, error)
}

// courseService is the implementation of CourseService
type courseService struct {
	repo repository.CourseRepository
}

// NewCourseService creates a new CourseService
func NewCourseService(repo repository.CourseRepository) CourseService {
	return &courseService{repo: repo}
}

// CreateCourse creates a new course record
func (s *courseService) CreateCourse(ctx context.Context, c *model.Course) (*model.Course, error) {
	err := s.repo.CreateCourse(ctx, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
