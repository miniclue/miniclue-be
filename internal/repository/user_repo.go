package repository

import (
	"app/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateAPIKeyFlag(ctx context.Context, userID string, provider string, hasKey bool) error
	UpdateModelPreference(ctx context.Context, userID string, provider string, model string, enabled bool) error
}

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepository {
	return &userRepo{pool: pool}
}

func (r *userRepo) CreateUser(ctx context.Context, u *model.User) error {
	// Initialize APIKeysProvided if nil
	if u.APIKeysProvided == nil {
		u.APIKeysProvided = make(model.APIKeysProvided)
	}
	if u.ModelPreferences == nil {
		u.ModelPreferences = make(model.ModelPreferences)
	}
	apiKeysJSON, err := json.Marshal(u.APIKeysProvided)
	if err != nil {
		return fmt.Errorf("marshaling API keys: %w", err)
	}
	modelPrefsJSON, err := json.Marshal(u.ModelPreferences)
	if err != nil {
		return fmt.Errorf("marshaling model preferences: %w", err)
	}

	query := `INSERT INTO user_profiles (user_id, name, email, avatar_url, api_keys_provided, model_preferences) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb) ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, avatar_url = EXCLUDED.avatar_url, api_keys_provided = EXCLUDED.api_keys_provided, model_preferences = EXCLUDED.model_preferences, updated_at = NOW() RETURNING user_id, name, email, avatar_url, api_keys_provided, model_preferences, created_at, updated_at;`
	err = r.pool.QueryRow(ctx, query, u.UserID, u.Name, u.Email, u.AvatarURL, apiKeysJSON, modelPrefsJSON).Scan(&u.UserID, &u.Name, &u.Email, &u.AvatarURL, &u.APIKeysProvided, &u.ModelPreferences, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating user %s: %w", u.UserID, err)
	}
	return nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	query := `SELECT user_id, email, name, avatar_url, api_keys_provided, model_preferences, created_at, updated_at FROM user_profiles WHERE user_id=$1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.UserID, &u.Email, &u.Name, &u.AvatarURL, &u.APIKeysProvided, &u.ModelPreferences, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by id %s: %w", id, err)
	}
	return &u, nil
}

func (r *userRepo) UpdateAPIKeyFlag(ctx context.Context, userID string, provider string, hasKey bool) error {
	// Use jsonb_set with to_jsonb to properly convert the boolean to JSONB
	// This avoids issues with JSON marshaling and type casting in PostgreSQL
	query := `UPDATE user_profiles SET api_keys_provided = jsonb_set(COALESCE(api_keys_provided, '{}'::jsonb), ARRAY[$1], to_jsonb($2::boolean), true), updated_at = NOW() WHERE user_id = $3`
	result, err := r.pool.Exec(ctx, query, provider, hasKey, userID)
	if err != nil {
		return fmt.Errorf("updating API key flag for user %s, provider %s: %w", userID, provider, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected: user %s may not exist in database", userID)
	}
	return nil
}

func (r *userRepo) UpdateModelPreference(ctx context.Context, userID string, provider string, modelName string, enabled bool) error {
	// First ensure the provider key exists with an empty object if it doesn't exist,
	// then set the model value within that provider
	// This handles the case where model_preferences is empty or the provider key doesn't exist
	query := `
		UPDATE user_profiles
		SET model_preferences = jsonb_set(
			COALESCE(model_preferences, '{}'::jsonb) || jsonb_build_object($1::text, COALESCE(model_preferences->$1, '{}'::jsonb)),
			ARRAY[$1::text, $2::text],
			to_jsonb($3::boolean),
			true
		),
		updated_at = NOW()
		WHERE user_id = $4
	`
	result, err := r.pool.Exec(ctx, query, provider, modelName, enabled, userID)
	if err != nil {
		return fmt.Errorf("updating model preference for user %s, provider %s, model %s: %w", userID, provider, modelName, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected: user %s may not exist in database", userID)
	}
	return nil
}
