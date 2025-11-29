package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app/internal/api/v1/dto"
	"app/internal/middleware"
	"app/internal/model"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type ChatHandler struct {
	chatService service.ChatService
	validate    *validator.Validate
	logger      zerolog.Logger
}

func NewChatHandler(chatService service.ChatService, validate *validator.Validate, logger zerolog.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		validate:    validate,
		logger:      logger,
	}
}

func (h *ChatHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	// ChatHandler routes are registered through LectureHandler to avoid route conflicts
	// Chat routes are handled via delegation from LectureHandler.handleLecture
}

func (h *ChatHandler) handleChatRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/lectures/") {
		http.NotFound(w, r)
		return
	}

	// Extract lecture ID and remaining path
	pathParts := strings.Split(strings.TrimPrefix(path, "/lectures/"), "/")
	if len(pathParts) < 2 {
		http.NotFound(w, r)
		return
	}

	lectureID := pathParts[0]
	remainingPath := strings.Join(pathParts[1:], "/")

	switch {
	case remainingPath == "chats" && r.Method == http.MethodPost:
		h.createChat(w, r, lectureID)
	case remainingPath == "chats" && r.Method == http.MethodGet:
		h.listChats(w, r, lectureID)
	case strings.HasPrefix(remainingPath, "chats/") && strings.HasSuffix(remainingPath, "/stream") && r.Method == http.MethodPost:
		chatID := strings.TrimSuffix(strings.TrimPrefix(remainingPath, "chats/"), "/stream")
		h.streamChat(w, r, lectureID, chatID)
	case strings.HasPrefix(remainingPath, "chats/") && strings.HasSuffix(remainingPath, "/messages") && r.Method == http.MethodGet:
		chatID := strings.TrimSuffix(strings.TrimPrefix(remainingPath, "chats/"), "/messages")
		h.listMessages(w, r, lectureID, chatID)
	case strings.HasPrefix(remainingPath, "chats/") && r.Method == http.MethodGet:
		chatID := strings.TrimPrefix(remainingPath, "chats/")
		h.getChat(w, r, lectureID, chatID)
	case strings.HasPrefix(remainingPath, "chats/") && r.Method == http.MethodDelete:
		chatID := strings.TrimPrefix(remainingPath, "chats/")
		h.deleteChat(w, r, lectureID, chatID)
	default:
		http.NotFound(w, r)
	}
}

// createChat godoc
// @Summary Create a new chat
// @Description Creates a new chat conversation for a lecture. The chat title is optional and defaults to "New Chat".
// @Tags chats
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param chat body dto.ChatCreateDTO false "Chat creation request"
// @Success 201 {object} dto.ChatResponseDTO
// @Failure 400 {string} string "Invalid JSON payload"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to create chat"
// @Router /lectures/{lectureId}/chats [post]
func (h *ChatHandler) createChat(w http.ResponseWriter, r *http.Request, lectureID string) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req dto.ChatCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	title := "New Chat"
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	}

	chat, err := h.chatService.CreateChat(r.Context(), lectureID, userID, title)
	if err != nil {
		if err == service.ErrUnauthorized || err == service.ErrLectureNotFound {
			http.Error(w, "Lecture not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to create chat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.ChatResponseDTO{
		ID:        chat.ID,
		LectureID: chat.LectureID,
		UserID:    chat.UserID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// listChats godoc
// @Summary List chats for a lecture
// @Description Retrieves all chats for a specific lecture with pagination support.
// @Tags chats
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param limit query int false "Maximum number of chats to return" default(50)
// @Param offset query int false "Number of chats to skip" default(0)
// @Success 200 {array} dto.ChatResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to list chats"
// @Router /lectures/{lectureId}/chats [get]
func (h *ChatHandler) listChats(w http.ResponseWriter, r *http.Request, lectureID string) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	chats, err := h.chatService.ListChats(r.Context(), lectureID, userID, limit, offset)
	if err != nil {
		if err == service.ErrUnauthorized || err == service.ErrLectureNotFound {
			http.Error(w, "Lecture not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to list chats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ChatResponseDTO, len(chats))
	for i, chat := range chats {
		resp[i] = dto.ChatResponseDTO{
			ID:        chat.ID,
			LectureID: chat.LectureID,
			UserID:    chat.UserID,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// getChat godoc
// @Summary Get a chat
// @Description Retrieves a specific chat by its ID.
// @Tags chats
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param chatId path string true "Chat ID"
// @Success 200 {object} dto.ChatResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Chat not found"
// @Failure 500 {string} string "Failed to get chat"
// @Router /lectures/{lectureId}/chats/{chatId} [get]
func (h *ChatHandler) getChat(w http.ResponseWriter, r *http.Request, lectureID, chatID string) {
	_ = lectureID
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	chat, err := h.chatService.GetChat(r.Context(), chatID, userID)
	if err != nil {
		if err == service.ErrChatNotFound {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get chat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.ChatResponseDTO{
		ID:        chat.ID,
		LectureID: chat.LectureID,
		UserID:    chat.UserID,
		Title:     chat.Title,
		CreatedAt: chat.CreatedAt,
		UpdatedAt: chat.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// deleteChat godoc
// @Summary Delete a chat
// @Description Deletes a chat and all its associated messages.
// @Tags chats
// @Param lectureId path string true "Lecture ID"
// @Param chatId path string true "Chat ID"
// @Success 204 {string} string "No Content"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Chat not found"
// @Failure 500 {string} string "Failed to delete chat"
// @Router /lectures/{lectureId}/chats/{chatId} [delete]
func (h *ChatHandler) deleteChat(w http.ResponseWriter, r *http.Request, lectureID, chatID string) {
	_ = lectureID
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	err := h.chatService.DeleteChat(r.Context(), chatID, userID)
	if err != nil {
		if err == service.ErrChatNotFound || err == service.ErrUnauthorized {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete chat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listMessages godoc
// @Summary List messages in a chat
// @Description Retrieves all messages for a specific chat with pagination support. Messages are returned in chronological order (oldest first).
// @Tags chats
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param chatId path string true "Chat ID"
// @Param limit query int false "Maximum number of messages to return" default(100)
// @Param offset query int false "Number of messages to skip" default(0)
// @Success 200 {array} dto.MessageResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Chat not found"
// @Failure 500 {string} string "Failed to list messages"
// @Router /lectures/{lectureId}/chats/{chatId}/messages [get]
func (h *ChatHandler) listMessages(w http.ResponseWriter, r *http.Request, lectureID, chatID string) {
	_ = lectureID
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	limit := 100
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := h.chatService.ListMessages(r.Context(), chatID, userID, limit, offset)
	if err != nil {
		if err == service.ErrChatNotFound || err == service.ErrUnauthorized {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to list messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]dto.MessageResponseDTO, len(messages))
	for i, msg := range messages {
		parts := make([]dto.MessagePartDTO, len(msg.Parts))
		for j, part := range msg.Parts {
			parts[j] = dto.MessagePartDTO{
				Type: part.Type,
				Text: part.Text,
			}
		}
		resp[i] = dto.MessageResponseDTO{
			ID:        msg.ID,
			ChatID:    msg.ChatID,
			Role:      msg.Role,
			Parts:     parts,
			CreatedAt: msg.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// streamChat godoc
// @Summary Stream chat response
// @Description Sends a user message and streams the AI assistant's response using Server-Sent Events (SSE). The user message is saved immediately, and the assistant response is saved after streaming completes. The model parameter specifies which LLM model to use for the response.
// @Tags chats
// @Accept json
// @Produce text/event-stream
// @Param lectureId path string true "Lecture ID"
// @Param chatId path string true "Chat ID"
// @Param request body dto.ChatStreamRequestDTO true "Chat stream request with message parts and model"
// @Success 200 {string} string "Server-Sent Events stream"
// @Failure 400 {string} string "Invalid JSON payload, validation failed, or API key required"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Chat or lecture not found"
// @Failure 500 {string} string "Failed to stream chat response"
// @Router /lectures/{lectureId}/chats/{chatId}/stream [post]
func (h *ChatHandler) streamChat(w http.ResponseWriter, r *http.Request, lectureID, chatID string) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req dto.ChatStreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Convert DTO to model
	messageParts := make(model.MessageParts, len(req.Parts))
	for i, part := range req.Parts {
		messageParts[i] = model.MessagePart{
			Type: part.Type,
			Text: part.Text,
		}
	}

	// Save user message first
	userMessage, err := h.chatService.CreateMessage(r.Context(), chatID, userID, "user", messageParts)
	if err != nil {
		if err == service.ErrChatNotFound || err == service.ErrUnauthorized {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to create message: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = userMessage // User message saved

	// Stream response from Python service
	stream, err := h.chatService.StreamChatResponse(r.Context(), lectureID, chatID, userID, messageParts, req.Model)
	if err != nil {
		if err == service.ErrChatNotFound || err == service.ErrUnauthorized || err == service.ErrLectureNotFound {
			http.Error(w, "Chat or lecture not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to stream chat response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			h.logger.Error().Err(err).Msg("Failed to close stream")
		}
	}()

	// Set SSE headers according to AI SDK Data Stream Protocol
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")             // Disable nginx buffering
	w.Header().Set("x-vercel-ai-ui-message-stream", "v1") // Required for AI SDK Data Stream Protocol

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Read from Python service stream and convert to AI SDK Data Stream Protocol
	reader := bufio.NewReader(stream)
	var fullContent strings.Builder
	// Generate a unique text part ID for this stream
	textPartID := fmt.Sprintf("part_%s_%d", chatID, time.Now().UnixNano())

	h.logger.Info().Str("chat_id", chatID).Str("lecture_id", lectureID).Msg("Starting to read from Python service stream")

	// Send text-start part (SDK will create the message internally)
	textStartPart := map[string]interface{}{
		"type": "text-start",
		"id":   textPartID,
	}
	textStartJSON, _ := json.Marshal(textStartPart)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", textStartJSON); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write text-start part")
		return
	}
	flusher.Flush()

	for {
		chunk, err := service.ParseSSEChunk(reader)
		if err != nil {
			if err == io.EOF {
				h.logger.Info().Msg("Reached end of Python service stream")
				break
			}
			h.logger.Error().Err(err).Msg("Error reading from Python service stream")
			break
		}

		h.logger.Debug().Interface("chunk", chunk).Msg("Received chunk from Python service")

		content, ok := chunk["content"].(string)
		if !ok {
			h.logger.Warn().Interface("chunk", chunk).Msg("Chunk missing or invalid content field, skipping")
			continue
		}

		done, _ := chunk["done"].(bool)

		h.logger.Debug().Str("content", content).Bool("done", done).Msg("Processing chunk")

		// Send text-delta parts according to AI SDK protocol
		if content != "" {
			deltaPart := map[string]interface{}{
				"type":  "text-delta",
				"id":    textPartID,
				"delta": content,
			}
			deltaJSON, err := json.Marshal(deltaPart)
			if err != nil {
				h.logger.Error().Err(err).Interface("deltaPart", deltaPart).Msg("Failed to marshal delta part")
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", deltaJSON); err != nil {
				h.logger.Error().Err(err).Msg("Failed to write delta part")
				break
			}
			flusher.Flush()
		}

		// Accumulate content for saving
		fullContent.WriteString(content)

		if done {
			h.logger.Info().Int("total_length", fullContent.Len()).Msg("Stream completed (done=true)")
			break
		}
	}

	// Send text-end part according to AI SDK protocol
	textEndPart := map[string]interface{}{
		"type": "text-end",
		"id":   textPartID,
	}
	textEndJSON, _ := json.Marshal(textEndPart)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", textEndJSON); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write text-end part")
	} else {
		flusher.Flush()
	}

	// Send finish part
	finishPart := map[string]interface{}{
		"type": "finish",
	}
	finishJSON, _ := json.Marshal(finishPart)
	if _, err := fmt.Fprintf(w, "data: %s\n\n", finishJSON); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write finish part")
	} else {
		flusher.Flush()
	}

	// Send [DONE] marker to terminate stream
	if _, err := fmt.Fprintf(w, "data: [DONE]\n\n"); err != nil {
		h.logger.Error().Err(err).Msg("Failed to write [DONE] marker")
	} else {
		flusher.Flush()
	}

	// Save assistant message
	if fullContent.Len() > 0 {
		contentStr := fullContent.String()
		h.logger.Info().Int("content_length", len(contentStr)).Str("chat_id", chatID).Msg("Saving assistant message to database")
		assistantParts := model.MessageParts{
			{Type: "text", Text: contentStr},
		}
		message, err := h.chatService.CreateMessage(r.Context(), chatID, userID, "assistant", assistantParts)
		if err != nil {
			h.logger.Error().Err(err).Str("chat_id", chatID).Msg("Failed to save assistant message")
		} else {
			h.logger.Info().Str("message_id", message.ID).Str("chat_id", chatID).Msg("Successfully saved assistant message")
		}
	} else {
		h.logger.Warn().Str("chat_id", chatID).Msg("No content to save for assistant message")
	}
}
