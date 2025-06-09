package service

import (
	"context"
	"errors"

	"app/internal/model"
	"app/internal/repository"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)

type UserService interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	GetRecentLectures(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
	GetCourses(ctx context.Context, userID string) ([]model.Course, error)
}

type userService struct {
	userRepo    repository.UserRepository
	courseRepo  repository.CourseRepository
	lectureRepo repository.LectureRepository
}

func NewUserService(userRepo repository.UserRepository, courseRepo repository.CourseRepository, lectureRepo repository.LectureRepository) UserService {
	return &userService{userRepo: userRepo, courseRepo: courseRepo, lectureRepo: lectureRepo}
}

func (s *userService) Create(ctx context.Context, u *model.User) (*model.User, error) {
	err := s.userRepo.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Get(ctx context.Context, id string) (*model.User, error) {
	u, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *userService) GetCourses(ctx context.Context, userID string) ([]model.Course, error) {
	courses, err := s.courseRepo.GetCoursesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (s *userService) GetRecentLectures(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error) {
	lectures, err := s.lectureRepo.GetLecturesByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return lectures, nil
}
