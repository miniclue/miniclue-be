package model

import "time"

// User represents a user in the system
type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"-"` // omit from JSON responses
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
