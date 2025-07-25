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
	GetRecentLectures(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error)
	GetCourses(ctx context.Context, userID string) ([]model.Course, error)
}

type userService struct {
	userRepo         repository.UserRepository
	courseRepo       repository.CourseRepository
	lectureRepo      repository.LectureRepository
	subscriptionRepo repository.SubscriptionRepository
	userLogger       zerolog.Logger
}

func NewUserService(userRepo repository.UserRepository, courseRepo repository.CourseRepository, lectureRepo repository.LectureRepository, subscriptionRepo repository.SubscriptionRepository, logger zerolog.Logger) UserService {
	// subscriptionRepo used to onboard new users to beta plan
	return &userService{
		userRepo:         userRepo,
		courseRepo:       courseRepo,
		lectureRepo:      lectureRepo,
		subscriptionRepo: subscriptionRepo,
		userLogger:       logger.With().Str("service", "UserService").Logger(),
	}
}

func (s *userService) Create(ctx context.Context, u *model.User) (*model.User, error) {
	err := s.userRepo.CreateUser(ctx, u)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to create user")
		return nil, err
	}
	// Onboard new user to default subscription (currently 'beta')
	if err := s.subscriptionRepo.UpsertSubscription(ctx, u.UserID, "beta"); err != nil {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to assign subscription")
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

func (s *userService) GetRecentLectures(ctx context.Context, userID string, limit, offset int) ([]model.Lecture, error) {
	lectures, err := s.lectureRepo.GetLecturesByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get recent lectures by user ID")
		return nil, err
	}
	return lectures, nil
}
