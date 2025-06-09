// internal/service/user_service.go
package service

import (
    "context"
    "errors"

    "app/internal/model"
    "app/internal/repository"
    "app/internal/util"
)

var (
    ErrUserNotFound          = errors.New("user not found")
    ErrEmailAlreadyRegistered = errors.New("email already registered")
)

type UserService interface {
    Register(ctx context.Context, u *model.User) (*model.User, error)
    Get(ctx context.Context, id int64) (*model.User, error)
}

type userService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, u *model.User) (*model.User, error) {
    // 1) Domain check
    exists, err := s.repo.GetByEmail(ctx, u.Email)
    if err != nil {
        return nil, err
    }
    if exists != nil {
        return nil, ErrEmailAlreadyRegistered
    }

    // 2) Hash password
    hash, err := util.HashPassword(u.Password)
    if err != nil {
        return nil, err
    }
    u.Password = hash

    // 3) Persist
    if err := s.repo.Create(ctx, u); err != nil {
        return nil, err
    }
    return u, nil
}

func (s *userService) Get(ctx context.Context, id int64) (*model.User, error) {
    u, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if u == nil {
        return nil, ErrUserNotFound
    }
    return u, nil
}
