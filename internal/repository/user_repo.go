package repository

import (
	"app/internal/model"
	"context"
	"database/sql"
	"errors"
	"time"
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
	query := `INSERT INTO user_profiles (email, name, avatar_url, created_at)
              VALUES ($1, $2, $3, $4) RETURNING user_id`
	now := time.Now().UTC()
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Name, u.AvatarURL, now).Scan(&u.UserID)
	if err != nil {
		return err
	}
	u.CreatedAt = now
	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	query := `SELECT user_id, email, name, avatar_url, created_at FROM user_profiles WHERE user_id=$1`
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}