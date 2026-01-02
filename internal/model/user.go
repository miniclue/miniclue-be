package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	UserID           string           `db:"user_id" json:"user_id"`
	Name             string           `db:"name" json:"name"`
	Email            string           `db:"email" json:"email"`
	AvatarURL        string           `db:"avatar_url" json:"avatar_url"`
	APIKeysProvided  APIKeysProvided  `db:"api_keys_provided" json:"api_keys_provided"`
	ModelPreferences ModelPreferences `db:"model_preferences" json:"model_preferences"`
	CreatedAt        time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at" json:"updated_at"`
}

// APIKeysProvided is a map of provider names to boolean flags
type APIKeysProvided map[string]bool

// Value implements the driver.Valuer interface for JSONB
func (a APIKeysProvided) Value() (driver.Value, error) {
	if a == nil {
		bytes, err := json.Marshal(map[string]bool{})
		if err != nil {
			return nil, err
		}
		return string(bytes), nil
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements the sql.Scanner interface for JSONB
func (a *APIKeysProvided) Scan(value interface{}) error {
	if value == nil {
		*a = make(map[string]bool)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*a = make(map[string]bool)
		return fmt.Errorf("cannot scan %T into APIKeysProvided", value)
	}

	if len(bytes) == 0 {
		*a = make(map[string]bool)
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// ModelPreferences stores per-provider model toggles
// Structure: provider -> model -> enabled
type ModelPreferences map[string]map[string]bool

// Value implements the driver.Valuer interface for JSONB
func (m ModelPreferences) Value() (driver.Value, error) {
	if m == nil {
		bytes, err := json.Marshal(map[string]map[string]bool{})
		if err != nil {
			return nil, err
		}
		return string(bytes), nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements the sql.Scanner interface for JSONB
func (m *ModelPreferences) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]map[string]bool)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*m = make(map[string]map[string]bool)
		return fmt.Errorf("cannot scan %T into ModelPreferences", value)
	}

	if len(bytes) == 0 {
		*m = make(map[string]map[string]bool)
		return nil
	}

	return json.Unmarshal(bytes, m)
}
