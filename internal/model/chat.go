package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Chat represents a chat conversation for a lecture
type Chat struct {
	ID        string    `db:"id" json:"id"`
	LectureID string    `db:"lecture_id" json:"lecture_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Title     string    `db:"title" json:"title"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Message represents a message in a chat (V2 format)
type Message struct {
	ID        string       `db:"id" json:"id"`
	ChatID    string       `db:"chat_id" json:"chat_id"`
	Role      string       `db:"role" json:"role"` // 'user' or 'assistant'
	Parts     MessageParts `db:"parts" json:"parts"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
}

// MessageParts is an array of message parts (JSONB)
type MessageParts []MessagePart

// MessagePart represents a single part of a message
type MessagePart struct {
	Type      string         `json:"type"` // 'text' or 'data-reference'
	Text      string         `json:"text,omitempty"`
	Reference *Reference     `json:"reference,omitempty"`
	Data      *ReferencePart `json:"data,omitempty"`
}

type ReferencePart struct {
	Type      string     `json:"type"`
	Text      string     `json:"text,omitempty"`
	Reference *Reference `json:"reference,omitempty"`
}

// Reference represents a contextual reference in a message part
type Reference struct {
	Type     string         `json:"type"` // e.g., 'slide'
	ID       string         `json:"id"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Value implements the driver.Valuer interface for JSONB
func (m MessageParts) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal([]MessagePart{})
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *MessageParts) Scan(value interface{}) error {
	if value == nil {
		*m = make(MessageParts, 0)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*m = make(MessageParts, 0)
		return fmt.Errorf("cannot scan %T into MessageParts", value)
	}

	if len(bytes) == 0 {
		*m = make(MessageParts, 0)
		return nil
	}

	return json.Unmarshal(bytes, m)
}
