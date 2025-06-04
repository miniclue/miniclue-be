package repository

import (
	"app/internal/model"
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// ... other methods (Update, Delete, List, etc.)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users (email, name, password, created_at)
              VALUES ($1, $2, $3, $4) RETURNING id`
	now := time.Now().UTC()
	err := r.db.QueryRowContext(ctx, query, u.Email, u.Name, u.Password, now).Scan(&u.ID)
	if err != nil {
		return err
	}
	u.CreatedAt = now
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	query := `SELECT id, email, name, password, created_at FROM users WHERE id=$1`
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// Similar query with WHERE email=$1
	// ...
	return nil, nil
}
