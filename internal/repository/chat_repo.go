package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"app/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	CreateChat(ctx context.Context, lectureID, userID, title string) (*model.Chat, error)
	GetChat(ctx context.Context, chatID, userID string) (*model.Chat, error)
	ListChats(ctx context.Context, lectureID, userID string, limit, offset int) ([]model.Chat, error)
	DeleteChat(ctx context.Context, chatID, userID string) error
	CreateMessage(ctx context.Context, chatID, role string, parts model.MessageParts) (*model.Message, error)
	ListMessages(ctx context.Context, chatID, userID string, limit, offset int) ([]model.Message, error)
}

type chatRepo struct {
	pool *pgxpool.Pool
}

func NewChatRepo(pool *pgxpool.Pool) ChatRepository {
	return &chatRepo{pool: pool}
}

func (r *chatRepo) CreateChat(ctx context.Context, lectureID, userID, title string) (*model.Chat, error) {
	query := `
		INSERT INTO chats (lecture_id, user_id, title)
		VALUES ($1, $2, $3)
		RETURNING id, lecture_id, user_id, title, created_at, updated_at
	`
	var chat model.Chat
	err := r.pool.QueryRow(ctx, query, lectureID, userID, title).Scan(
		&chat.ID,
		&chat.LectureID,
		&chat.UserID,
		&chat.Title,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating chat: %w", err)
	}
	return &chat, nil
}

func (r *chatRepo) GetChat(ctx context.Context, chatID, userID string) (*model.Chat, error) {
	query := `
		SELECT id, lecture_id, user_id, title, created_at, updated_at
		FROM chats
		WHERE id = $1 AND user_id = $2
	`
	var chat model.Chat
	err := r.pool.QueryRow(ctx, query, chatID, userID).Scan(
		&chat.ID,
		&chat.LectureID,
		&chat.UserID,
		&chat.Title,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("chat not found: %w", err)
		}
		return nil, fmt.Errorf("getting chat: %w", err)
	}
	return &chat, nil
}

func (r *chatRepo) ListChats(ctx context.Context, lectureID, userID string, limit, offset int) ([]model.Chat, error) {
	query := fmt.Sprintf(`
		SELECT id, lecture_id, user_id, title, created_at, updated_at
		FROM chats
		WHERE lecture_id = $1 AND user_id = $2
		ORDER BY updated_at DESC
		LIMIT %d OFFSET %d
	`, limit, offset)

	rows, err := r.pool.Query(ctx, query, lectureID, userID)
	if err != nil {
		return nil, fmt.Errorf("querying chats: %w", err)
	}
	defer rows.Close()

	var chats []model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(
			&chat.ID,
			&chat.LectureID,
			&chat.UserID,
			&chat.Title,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning chat row: %w", err)
		}
		chats = append(chats, chat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating chat rows: %w", err)
	}

	return chats, nil
}

func (r *chatRepo) DeleteChat(ctx context.Context, chatID, userID string) error {
	query := `
		DELETE FROM chats
		WHERE id = $1 AND user_id = $2
	`
	result, err := r.pool.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("deleting chat: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("chat not found or access denied")
	}
	return nil
}

func (r *chatRepo) CreateMessage(ctx context.Context, chatID, role string, parts model.MessageParts) (*model.Message, error) {
	partsJSON, err := json.Marshal(parts)
	if err != nil {
		return nil, fmt.Errorf("marshaling message parts: %w", err)
	}

	query := `
		INSERT INTO messages (chat_id, role, parts)
		VALUES ($1, $2, $3::jsonb)
		RETURNING id, chat_id, role, parts, created_at
	`
	var message model.Message
	err = r.pool.QueryRow(ctx, query, chatID, role, partsJSON).Scan(
		&message.ID,
		&message.ChatID,
		&message.Role,
		&message.Parts,
		&message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("creating message: %w", err)
	}
	return &message, nil
}

func (r *chatRepo) ListMessages(ctx context.Context, chatID, userID string, limit, offset int) ([]model.Message, error) {
	// Verify chat ownership first
	chatQuery := `SELECT id FROM chats WHERE id = $1 AND user_id = $2`
	var chatIDCheck string
	err := r.pool.QueryRow(ctx, chatQuery, chatID, userID).Scan(&chatIDCheck)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("chat not found or access denied")
		}
		return nil, fmt.Errorf("verifying chat ownership: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, chat_id, role, parts, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT %d OFFSET %d
	`, limit, offset)

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.Role,
			&message.Parts,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating message rows: %w", err)
	}

	return messages, nil
}
