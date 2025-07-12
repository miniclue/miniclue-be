package handler

import (
	"encoding/json"
	"net/http"

	"app/internal/api/v1/dto"
	"app/internal/service"

	"github.com/rs/zerolog"
)

// DLQHandler handles dead-letter queue push events
type DLQHandler struct {
	service service.DLQService
	logger  zerolog.Logger
}

func NewDLQHandler(s service.DLQService, l zerolog.Logger) *DLQHandler {
	return &DLQHandler{service: s, logger: l}
}

// RegisterRoutes mounts the DLQ handler.
// This route is public and does not use the auth middleware.
func (h *DLQHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("POST /dlq", authMw(http.HandlerFunc(h.HandleDLQ)))
}

// HandleDLQ godoc
// @Summary Process dead-letter queue message
// @Description Receives a Pub/Sub push from a dead-letter topic and persists the message to the database for manual inspection.
// @Tags dlq
// @Accept json
// @Produce json
// @Param request body dto.PubSubPushRequest true "Dead-letter queue Pub/Sub push payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Invalid request body or format"
// @Failure 500 {string} string "Internal server error"
// @Router /dlq [post]
func (h *DLQHandler) HandleDLQ(w http.ResponseWriter, r *http.Request) {
	var req dto.PubSubPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode DLQ request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// The actual message is nested and needs to be checked
	if req.Message.MessageID == "" {
		h.logger.Error().Msg("Received empty or invalid DLQ message")
		http.Error(w, "Invalid Pub/Sub message format", http.StatusBadRequest)
		return
	}

	h.logger.Info().
		Str("messageId", req.Message.MessageID).
		Str("subscription", req.Subscription).
		Msg("Processing dead-letter queue message")

	if err := h.service.ProcessAndSave(r.Context(), &req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to save DLQ message to database")
		// Still return 2xx to Pub/Sub to prevent retries of a message that is already in the DLQ.
		// The error is logged for offline analysis.
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.logger.Info().
		Str("messageId", req.Message.MessageID).
		Msg("Successfully processed and saved DLQ message")

	w.WriteHeader(http.StatusNoContent)
}
