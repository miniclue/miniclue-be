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
	GetUsage(ctx context.Context, userID string) (*model.UserUsage, error)
}

type userService struct {
	userRepo         repository.UserRepository
	courseRepo       repository.CourseRepository
	lectureRepo      repository.LectureRepository
	subscriptionRepo repository.SubscriptionRepository
	stripeSvc        *StripeService
	userLogger       zerolog.Logger
}

func NewUserService(userRepo repository.UserRepository, courseRepo repository.CourseRepository, lectureRepo repository.LectureRepository, subscriptionRepo repository.SubscriptionRepository, stripeSvc *StripeService, logger zerolog.Logger) UserService {
	// subscriptionRepo used to onboard new users to beta plan
	return &userService{
		userRepo:         userRepo,
		courseRepo:       courseRepo,
		lectureRepo:      lectureRepo,
		subscriptionRepo: subscriptionRepo,
		stripeSvc:        stripeSvc,
		userLogger:       logger.With().Str("service", "UserService").Logger(),
	}
}

func (s *userService) Create(ctx context.Context, u *model.User) (*model.User, error) {
	// Create user in database first
	err := s.userRepo.CreateUser(ctx, u)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", u.UserID).Msg("Failed to create user")
		return nil, err
	}

	// Create Stripe customer (non-blocking - if it fails, user can still use the app)
	if s.stripeSvc != nil {
		customerID, err := s.stripeSvc.CreateCustomer(ctx, u)
		if err != nil {
			s.userLogger.Warn().Err(err).Str("user_id", u.UserID).Msg("Failed to create Stripe customer during signup - user can still use the app")
			// Don't return error - user can still use the app without Stripe customer
		} else {
			s.userLogger.Info().Str("user_id", u.UserID).Str("stripe_customer_id", customerID).Msg("Created Stripe customer during signup")
		}
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

func (s *userService) GetUsage(ctx context.Context, userID string) (*model.UserUsage, error) {
	usage, err := s.userRepo.GetUserUsage(ctx, userID)
	if err != nil {
		s.userLogger.Error().Err(err).Str("user_id", userID).Msg("Failed to get usage for user")
		return nil, err
	}
	return usage, nil
}
