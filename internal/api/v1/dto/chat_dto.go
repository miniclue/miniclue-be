package dto

import "time"

type ChatCreateDTO struct {
	Title *string `json:"title,omitempty"`
}

type ChatUpdateDTO struct {
	Title *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
}

type ChatResponseDTO struct {
	ID        string    `json:"id"`
	LectureID string    `json:"lecture_id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessagePartDTO struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type MessageCreateDTO struct {
	Parts []MessagePartDTO `json:"parts" validate:"required"`
}

type MessageResponseDTO struct {
	ID        string           `json:"id"`
	ChatID    string           `json:"chat_id"`
	Role      string           `json:"role"`
	Parts     []MessagePartDTO `json:"parts"`
	CreatedAt time.Time        `json:"created_at"`
}

type ChatStreamRequestDTO struct {
	Parts []MessagePartDTO `json:"parts" validate:"required"`
	Model string           `json:"model" validate:"required"`
}
