package repository

import (
	"app/internal/model"
	"context"
	"database/sql"
	"errors"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(ctx context.Context, u *model.User) error {
	query := `INSERT INTO user_profiles (user_id, name, email, avatar_url) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, avatar_url = EXCLUDED.avatar_url, updated_at = NOW() RETURNING user_id, name, email, avatar_url, created_at, updated_at;`
	err := r.db.QueryRowContext(ctx, query, u.UserID, u.Name, u.Email, u.AvatarURL).Scan(&u.UserID, &u.Name, &u.Email, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	query := `SELECT user_id, email, name, avatar_url, created_at, updated_at FROM user_profiles WHERE user_id=$1`
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
