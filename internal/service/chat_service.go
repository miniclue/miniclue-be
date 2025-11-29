package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"app/internal/model"
	"app/internal/repository"

	"github.com/rs/zerolog"
)

var (
	ErrChatNotFound    = errors.New("chat not found")
	ErrLectureNotFound = errors.New("lecture not found")
	ErrUnauthorized    = errors.New("unauthorized access")
)

type ChatService interface {
	CreateChat(ctx context.Context, lectureID, userID, title string) (*model.Chat, error)
	GetChat(ctx context.Context, chatID, userID string) (*model.Chat, error)
	ListChats(ctx context.Context, lectureID, userID string, limit, offset int) ([]model.Chat, error)
	UpdateChat(ctx context.Context, chatID, userID, title string) (*model.Chat, error)
	DeleteChat(ctx context.Context, chatID, userID string) error
	CreateMessage(ctx context.Context, chatID, userID, role string, parts model.MessageParts) (*model.Message, error)
	ListMessages(ctx context.Context, chatID, userID string, limit int) ([]model.Message, error)
	GetMessageCount(ctx context.Context, chatID, userID string) (int, error)
	StreamChatResponse(ctx context.Context, lectureID, chatID, userID string, messageParts model.MessageParts, model string) (io.ReadCloser, error)
	GenerateAndUpdateTitle(ctx context.Context, lectureID, chatID, userID string, userMessageParts model.MessageParts, assistantMessageParts model.MessageParts)
}

type chatService struct {
	chatRepo     repository.ChatRepository
	lectureRepo  repository.LectureRepository
	pythonClient PythonClient
	logger       zerolog.Logger
}

func NewChatService(
	chatRepo repository.ChatRepository,
	lectureRepo repository.LectureRepository,
	pythonClient PythonClient,
	logger zerolog.Logger,
) ChatService {
	return &chatService{
		chatRepo:     chatRepo,
		lectureRepo:  lectureRepo,
		pythonClient: pythonClient,
		logger:       logger.With().Str("service", "ChatService").Logger(),
	}
}

func (s *chatService) CreateChat(ctx context.Context, lectureID, userID, title string) (*model.Chat, error) {
	// Verify lecture exists and user owns it
	lecture, err := s.lectureRepo.GetLectureByID(ctx, lectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return nil, ErrUnauthorized
	}

	if title == "" {
		title = "New Chat"
	}

	chat, err := s.chatRepo.CreateChat(ctx, lectureID, userID, title)
	if err != nil {
		s.logger.Error().Err(err).Str("lecture_id", lectureID).Str("user_id", userID).Msg("Failed to create chat")
		return nil, fmt.Errorf("creating chat: %w", err)
	}

	return chat, nil
}

func (s *chatService) GetChat(ctx context.Context, chatID, userID string) (*model.Chat, error) {
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting chat: %w", err)
	}
	return chat, nil
}

func (s *chatService) ListChats(ctx context.Context, lectureID, userID string, limit, offset int) ([]model.Chat, error) {
	// Verify lecture exists and user owns it
	lecture, err := s.lectureRepo.GetLectureByID(ctx, lectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return nil, ErrUnauthorized
	}

	chats, err := s.chatRepo.ListChats(ctx, lectureID, userID, limit, offset)
	if err != nil {
		s.logger.Error().Err(err).Str("lecture_id", lectureID).Msg("Failed to list chats")
		return nil, fmt.Errorf("listing chats: %w", err)
	}

	return chats, nil
}

func (s *chatService) UpdateChat(ctx context.Context, chatID, userID, title string) (*model.Chat, error) {
	// Verify chat ownership
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	// Verify lecture ownership
	lecture, err := s.lectureRepo.GetLectureByID(ctx, chat.LectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return nil, ErrUnauthorized
	}

	updatedChat, err := s.chatRepo.UpdateChat(ctx, chatID, userID, title)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to update chat")
		return nil, fmt.Errorf("updating chat: %w", err)
	}

	return updatedChat, nil
}

func (s *chatService) DeleteChat(ctx context.Context, chatID, userID string) error {
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("chat not found: %w", err)
	}

	// Verify lecture ownership
	lecture, err := s.lectureRepo.GetLectureByID(ctx, chat.LectureID)
	if err != nil {
		return fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return ErrUnauthorized
	}

	err = s.chatRepo.DeleteChat(ctx, chatID, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to delete chat")
		return fmt.Errorf("deleting chat: %w", err)
	}

	return nil
}

func (s *chatService) CreateMessage(ctx context.Context, chatID, userID, role string, parts model.MessageParts) (*model.Message, error) {
	// Verify chat ownership
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	// Verify lecture ownership
	lecture, err := s.lectureRepo.GetLectureByID(ctx, chat.LectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return nil, ErrUnauthorized
	}

	message, err := s.chatRepo.CreateMessage(ctx, chatID, role, parts)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to create message")
		return nil, fmt.Errorf("creating message: %w", err)
	}

	return message, nil
}

func (s *chatService) ListMessages(ctx context.Context, chatID, userID string, limit int) ([]model.Message, error) {
	// Verify chat ownership
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	// Verify lecture ownership
	lecture, err := s.lectureRepo.GetLectureByID(ctx, chat.LectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID {
		return nil, ErrUnauthorized
	}

	messages, err := s.chatRepo.ListMessages(ctx, chatID, userID, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to list messages")
		return nil, fmt.Errorf("listing messages: %w", err)
	}

	return messages, nil
}

func (s *chatService) GetMessageCount(ctx context.Context, chatID, userID string) (int, error) {
	count, err := s.chatRepo.GetMessageCount(ctx, chatID, userID)
	if err != nil {
		return 0, fmt.Errorf("getting message count: %w", err)
	}
	return count, nil
}

func (s *chatService) GenerateAndUpdateTitle(ctx context.Context, lectureID, chatID, userID string, userMessageParts model.MessageParts, assistantMessageParts model.MessageParts) {
	// Convert user message parts to map for Python service
	userMessagePartsMap := make([]map[string]interface{}, len(userMessageParts))
	for i, part := range userMessageParts {
		userMessagePartsMap[i] = map[string]interface{}{
			"type": part.Type,
			"text": part.Text,
		}
	}

	// Convert assistant message parts to map for Python service
	assistantMessagePartsMap := make([]map[string]interface{}, len(assistantMessageParts))
	for i, part := range assistantMessageParts {
		assistantMessagePartsMap[i] = map[string]interface{}{
			"type": part.Type,
			"text": part.Text,
		}
	}

	// Generate title via Python service
	title, err := s.pythonClient.GenerateChatTitle(ctx, lectureID, chatID, userID, userMessagePartsMap, assistantMessagePartsMap)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to generate chat title")
		return
	}

	// Update chat title in database - frontend will poll to get the updated title
	_, err = s.UpdateChat(ctx, chatID, userID, title)
	if err != nil {
		s.logger.Error().Err(err).Str("chat_id", chatID).Str("title", title).Msg("Failed to update chat title")
		return
	}
}

func (s *chatService) StreamChatResponse(ctx context.Context, lectureID, chatID, userID string, messageParts model.MessageParts, model string) (io.ReadCloser, error) {
	// Verify chat ownership
	chat, err := s.chatRepo.GetChat(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("chat not found: %w", err)
	}

	// Verify lecture ownership
	lecture, err := s.lectureRepo.GetLectureByID(ctx, lectureID)
	if err != nil {
		return nil, fmt.Errorf("lecture not found: %w", err)
	}
	if lecture.UserID != userID || chat.LectureID != lectureID {
		return nil, ErrUnauthorized
	}

	// Convert message parts to map for JSON serialization
	messagePartsMap := make([]map[string]interface{}, len(messageParts))
	for i, part := range messageParts {
		messagePartsMap[i] = map[string]interface{}{
			"type": part.Type,
			"text": part.Text,
		}
	}

	// Stream from Python service (Python will retrieve API key)
	stream, err := s.pythonClient.StreamChat(ctx, lectureID, chatID, userID, messagePartsMap, model)
	if err != nil {
		s.logger.Error().Err(err).Str("lecture_id", lectureID).Str("chat_id", chatID).Msg("Failed to stream chat response")
		return nil, fmt.Errorf("streaming chat response: %w", err)
	}

	return stream, nil
}
