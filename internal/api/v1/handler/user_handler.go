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
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{userService: userService, validate: v}
}

// RegisterRoutes mounts v1 user routes
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/users/me", authMw(http.HandlerFunc(h.handleUsers)))
	mux.Handle("/users/me/courses", authMw(http.HandlerFunc(h.getUserCourses)))
	mux.Handle("/users/me/recents", authMw(http.HandlerFunc(h.getRecentLectures)))
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
// @Summary Create a new user profile
// @Description Creates a new user profile associated with the authenticated user ID.
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
		UserID:    createdUser.UserID,
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		AvatarURL: createdUser.AvatarURL,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}

	// 7. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
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
		UserID:    user.UserID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getUserCourses godoc
// @Summary Get user's courses
// @Description Retrieves the list of courses associated with the authenticated user.
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
		})
	}

	// 4. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courseDTOs)
}

// getRecentLectures godoc
// @Summary Get recent lectures
// @Description Retrieves a list of recently viewed lectures for the authenticated user.
// @Tags users
// @Produce json
// @Param limit query int false "Number of lectures to return (default 10)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} model.Lecture
// @Failure 401 {string} string "Unauthorized: user ID not found in context"
// @Failure 500 {string} string "Failed to retrieve recent lectures"
// @Router /users/me/recents [get]
func (h *UserHandler) getRecentLectures(w http.ResponseWriter, r *http.Request) {
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

	// 3. Call service to get recent lectures
	lectures, err := h.userService.GetRecentLectures(r.Context(), userID, limit, offset)
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
		})
	}

	// 5. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lectureDTOs)
}
