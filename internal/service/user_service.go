package service

import (
	"context"
	"errors"

	"app/internal/model"
	"app/internal/repository"

	"github.com/rs/zerolog"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)

type UserService interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	GetRecentLecturesWithCount(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, int, error)
	GetCourses(ctx context.Context, userID string) ([]model.Course, error)
}

type userService struct {
	userRepo    repository.UserRepository
	courseRepo  repository.CourseRepository
	lectureRepo repository.LectureRepository
	userLogger  zerolog.Logger
}

func NewUserService(userRepo repository.UserRepository, courseRepo repository.CourseRepository, lectureRepo repository.LectureRepository, logger zerolog.Logger) UserService {
	return &userService{
		userRepo:    userRepo,
		courseRepo:  courseRepo,
		lectureRepo: lectureRepo,
		userLogger:  logger.With().Str("service", "UserService").Logger(),
	}
}

func (s *userService) Create(ctx context.Context, u *model.User) (*model.User, error) {
	// Check if user already exists first
	_, err := s.userRepo.GetUserByID(ctx, u.UserID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to check if user exists")
		return nil, err
	}

	// Create/update user in database
	err = s.userRepo.CreateUser(ctx, u)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to create/update user")
		return nil, err
	}

	return u, nil
}

func (s *userService) Get(ctx context.Context, id string) (*model.User, error) {
	u, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", id).Msg("Failed to get user by ID")
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
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get courses by user ID")
		return nil, err
	}
	return courses, nil
}

func (s *userService) GetRecentLecturesWithCount(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, int, error) {
	// Get lectures with pagination
	lectures, err := s.lectureRepo.GetLecturesByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get recent lectures by user ID")
		return nil, 0, err
	}

	// Get total count
	totalCount, err := s.lectureRepo.CountLecturesByUserID(ctx, userID)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get lecture count by user ID")
		return nil, 0, err
	}

	return lectures, totalCount, nil
}
