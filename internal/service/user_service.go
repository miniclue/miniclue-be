package service

import (
	"app/internal/model"
	"app/internal/repository"
	"app/internal/util" // e.g., for password hashing
	"context"
	"errors"
)

type UserService interface {
	Register(ctx context.Context, dto *model.UserCreateDTO) (*model.UserResponseDTO, error)
	Get(ctx context.Context, id int64) (*model.UserResponseDTO, error)
	// ... other methods
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, dto *model.UserCreateDTO) (*model.UserResponseDTO, error) {
	// 1. Validate DTO (can use validator in handler before calling service)
	// 2. Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}
	// 3. Hash password
	hashed, err := util.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}
	// 4. Create model.User
	user := &model.User{
		Email:    dto.Email,
		Name:     dto.Name,
		Password: hashed,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	// 5. Build response DTO
	resp := &model.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}
	return resp, nil
}

func (s *userService) Get(ctx context.Context, id int64) (*model.UserResponseDTO, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // or custom ErrNotFound
	}
	return &model.UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}
