package repository

import (
	"app/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID, customerID string) error
	// GetUserByStripeCustomerID returns the user associated with the given Stripe customer ID, or nil if none
	GetUserByStripeCustomerID(ctx context.Context, customerID string) (*model.User, error)
}

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) CreateUser(ctx context.Context, u *model.User) error {
	query := `INSERT INTO user_profiles (user_id, name, email, avatar_url) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, avatar_url = EXCLUDED.avatar_url, updated_at = NOW() RETURNING user_id, name, email, avatar_url, created_at, updated_at;`
	err := r.pool.QueryRow(ctx, query, u.UserID, u.Name, u.Email, u.AvatarURL).Scan(&u.UserID, &u.Name, &u.Email, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating user %s: %w", u.UserID, err)
	}
	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	query := `SELECT user_id, email, name, avatar_url, stripe_customer_id, created_at, updated_at FROM user_profiles WHERE user_id=$1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.UserID, &u.Email, &u.Name, &u.AvatarURL, &u.StripeCustomerID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by id %s: %w", id, err)
	}
	return &u, nil
}

func (r *userRepo) UpdateStripeCustomerID(ctx context.Context, userID, customerID string) error {
	const q = `UPDATE user_profiles SET stripe_customer_id = $2 WHERE user_id = $1`
	if _, err := r.pool.Exec(ctx, q, userID, customerID); err != nil {
		return fmt.Errorf("update stripe customer id for user %s: %w", userID, err)
	}
	return nil
}

// GetUserByStripeCustomerID returns the user whose stripe_customer_id matches the given ID.
func (r *userRepo) GetUserByStripeCustomerID(ctx context.Context, customerID string) (*model.User, error) {
	var u model.User
	const q = `SELECT user_id, email, name, avatar_url, stripe_customer_id, created_at, updated_at FROM user_profiles WHERE stripe_customer_id = $1`
	err := r.pool.QueryRow(ctx, q, customerID).Scan(
		&u.UserID,
		&u.Email,
		&u.Name,
		&u.AvatarURL,
		&u.StripeCustomerID,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by stripe customer id: %w", err)
	}
	return &u, nil
}
