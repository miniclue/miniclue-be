package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"app/internal/api/v1/dto"
	"app/internal/middleware"
	"app/internal/model"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
	logger      zerolog.Logger
}

func NewUserHandler(userService service.UserService, v *validator.Validate, logger zerolog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    v,
		logger:      logger,
	}
}

// RegisterRoutes mounts v1 user routes
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/users/me", authMw(http.HandlerFunc(h.handleUsers)))
	mux.Handle("/users/me/courses", authMw(http.HandlerFunc(h.getUserCourses)))
	mux.Handle("/users/me/recents", authMw(http.HandlerFunc(h.getRecentLecturesWithCount)))
	mux.Handle("/users/me/api-key", authMw(http.HandlerFunc(h.handleAPIKey)))
}

func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/users/me":
		h.createUser(w, r)

	case r.Method == http.MethodGet && r.URL.Path == "/users/me":
		h.getUser(w, r)
	default:
		http.NotFound(w, r)
	}
}

// createUser godoc
// @Summary Create or update a user profile
// @Description Creates a new user profile or updates an existing one associated with the authenticated user ID.
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.UserCreateDTO true "User creation request"
// @Success 201 {object} dto.UserResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 500 {string} string "Failed to create user"
// @Router /users/me [post]
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	// 1. Extract UserID from context
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userId == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// 2. Decode request body into DTO
	var req dto.UserCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Validate DTO
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Create model.User from DTO and context UserID
	userModel := &model.User{
		UserID:    userId,
		Name:      req.Name,
		Email:     req.Email,
		AvatarURL: req.AvatarURL,
	}

	// 5. Call service to create user profile
	createdUser, err := h.userService.Create(r.Context(), userModel)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Map domain model to response DTO
	resp := dto.UserResponseDTO{
		UserID:          createdUser.UserID,
		Name:            createdUser.Name,
		Email:           createdUser.Email,
		AvatarURL:       createdUser.AvatarURL,
		APIKeysProvided: createdUser.APIKeysProvided,
		CreatedAt:       createdUser.CreatedAt,
		UpdatedAt:       createdUser.UpdatedAt,
	}

	// 7. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Error already handled by http.Error in other cases
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getUser godoc
// @Summary Get user profile
// @Description Retrieves the profile of the authenticated user.
// @Tags users
// @Produce json
// @Success 200 {object} dto.UserResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/me [get]
func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.Get(r.Context(), userId)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resp := dto.UserResponseDTO{
		UserID:          user.UserID,
		Name:            user.Name,
		Email:           user.Email,
		AvatarURL:       user.AvatarURL,
		APIKeysProvided: user.APIKeysProvided,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Error already handled by http.Error in other cases
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getUserCourses godoc
// @Summary Get user's courses
// @Description Retrieves the list of courses associated with the authenticated user, sorted by most recent update.
// @Tags users
// @Produce json
// @Success 200 {array} dto.UserCourseResponseDTO
// @Failure 401 {string} string "Unauthorized: user ID not found in context"
// @Failure 500 {string} string "Failed to retrieve user courses"
// @Router /users/me/courses [get]
func (h *UserHandler) getUserCourses(w http.ResponseWriter, r *http.Request) {
	// 1. Check method
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	// 1. Extract UserID from context
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: user ID not found in context", http.StatusUnauthorized)
		return
	}

	// 2. Call service to get courses by user ID
	courses, err := h.userService.GetCourses(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user courses: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Map domain models to response DTOs
	var courseDTOs []dto.UserCourseResponseDTO
	for _, course := range courses {
		courseDTOs = append(courseDTOs, dto.UserCourseResponseDTO{
			CourseID:    course.CourseID,
			Title:       course.Title,
			Description: course.Description,
			IsDefault:   course.IsDefault,
			UpdatedAt:   course.UpdatedAt,
		})
	}

	// 4. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(courseDTOs); err != nil {
		// Error already handled by http.Error in other cases
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getRecentLecturesWithCount godoc
// @Summary Get recent lectures with count
// @Description Retrieves a list of recently viewed lectures for the authenticated user with total count.
// @Tags users
// @Produce json
// @Param limit query int false "Number of lectures to return (default 10)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {object} dto.UserRecentLecturesResponseDTO
// @Failure 401 {string} string "Unauthorized: user ID not found in context"
// @Failure 500 {string} string "Failed to retrieve recent lectures"
// @Router /users/me/recents [get]
func (h *UserHandler) getRecentLecturesWithCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Extract UserID from context
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: user ID not found in context", http.StatusUnauthorized)
		return
	}

	// 2. Parse limit and offset from query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// 3. Call service to get recent lectures with count
	lectures, totalCount, err := h.userService.GetRecentLecturesWithCount(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve recent lectures: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Map domain models to response DTOs
	var lectureDTOs []dto.UserRecentLectureResponseDTO
	for _, lecture := range lectures {
		lectureDTOs = append(lectureDTOs, dto.UserRecentLectureResponseDTO{
			LectureID: lecture.ID,
			Title:     lecture.Title,
			CourseID:  lecture.CourseID,
		})
	}

	// 5. Create response with lectures and total count
	response := dto.UserRecentLecturesResponseDTO{
		Lectures:   lectureDTOs,
		TotalCount: totalCount,
	}

	// 6. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Error already handled by http.Error in other cases
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleAPIKey handles API key operations (POST for storage, DELETE for removal)
func (h *UserHandler) handleAPIKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.storeAPIKey(w, r)
	case http.MethodDelete:
		h.deleteAPIKey(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// storeAPIKey godoc
// @Summary Store user's OpenAI API key
// @Description Stores the user's OpenAI API key securely in Google Cloud Secret Manager and updates the user profile flag.
// @Tags users
// @Accept json
// @Produce json
// @Param api_key body dto.APIKeyRequestDTO true "API key request"
// @Success 200 {object} dto.APIKeyResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 500 {string} string "Failed to store API key"
// @Router /users/me/api-key [post]
func (h *UserHandler) storeAPIKey(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userId == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req dto.APIKeyRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	err := h.userService.StoreAPIKey(r.Context(), userId, req.Provider, req.APIKey)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userId).Str("provider", req.Provider).Msg("Failed to store API key")
		http.Error(w, "Failed to store API key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.APIKeyResponseDTO{
		Provider:       req.Provider,
		HasProvidedKey: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// deleteAPIKey godoc
// @Summary Delete user's API key
// @Description Deletes the user's API key from Google Cloud Secret Manager and updates the user profile flag.
// @Tags users
// @Produce json
// @Param provider query string true "API provider (openai or gemini)"
// @Success 200 {object} dto.APIKeyResponseDTO
// @Failure 400 {string} string "Invalid provider parameter"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 500 {string} string "Failed to delete API key"
// @Router /users/me/api-key [delete]
func (h *UserHandler) deleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userId == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		http.Error(w, "Provider parameter is required", http.StatusBadRequest)
		return
	}

	if provider != "openai" && provider != "gemini" {
		http.Error(w, "Invalid provider. Must be 'openai' or 'gemini'", http.StatusBadRequest)
		return
	}

	err := h.userService.DeleteAPIKey(r.Context(), userId, provider)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userId).Str("provider", provider).Msg("Failed to delete API key")
		http.Error(w, "Failed to delete API key: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.APIKeyResponseDTO{
		Provider:       provider,
		HasProvidedKey: false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
