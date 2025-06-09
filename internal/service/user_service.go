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
	// RegisterUser(ctx context.Context, u *model.User) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// func (s *userService) RegisterUser(ctx context.Context, u *model.User) (*model.User, error) {
//     return nil, nil
// }

func (s *userService) GetUser(ctx context.Context, id string) (*model.User, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}
