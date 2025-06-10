package service

import (
	"context"

	"app/internal/model"
	"app/internal/repository"
)

// CourseService defines the interface for course operations
type CourseService interface {
	CreateCourse(ctx context.Context, c *model.Course) (*model.Course, error)
	// GetCourseByID retrieves a course by its ID
	GetCourseByID(ctx context.Context, courseID string) (*model.Course, error)
	// UpdateCourse updates an existing course
	UpdateCourse(ctx context.Context, c *model.Course) (*model.Course, error)
	// DeleteCourse deletes a course by its ID
	DeleteCourse(ctx context.Context, courseID string) error
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

// GetCourseByID retrieves a course by its ID
func (s *courseService) GetCourseByID(ctx context.Context, courseID string) (*model.Course, error) {
	return s.repo.GetCourseByID(ctx, courseID)
}

// UpdateCourse updates an existing course record
func (s *courseService) UpdateCourse(ctx context.Context, c *model.Course) (*model.Course, error) {
	if err := s.repo.UpdateCourse(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteCourse deletes a course by its ID
func (s *courseService) DeleteCourse(ctx context.Context, courseID string) error {
	return s.repo.DeleteCourse(ctx, courseID)
}
