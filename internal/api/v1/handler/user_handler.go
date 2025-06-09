package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"app/internal/api/v1/dto"
	"app/internal/middleware"

	"app/internal/model"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService   service.UserService
	courseService service.CourseService
	validate      *validator.Validate
}

func NewUserHandler(userService service.UserService, courseService service.CourseService, v *validator.Validate) *UserHandler {
	return &UserHandler{userService: userService, courseService: courseService, validate: v}
}

// RegisterRoutes mounts v1 user routes
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/users/me", authMw(http.HandlerFunc(h.handleUsers)))
	mux.Handle("/users/me/courses", authMw(http.HandlerFunc(h.getUserCourses)))
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
	createdUser, err := h.userService.CreateUser(r.Context(), userModel)
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

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUser(r.Context(), userId)
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
	courses, err := h.courseService.GetCoursesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user courses: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Map domain models to response DTOs
	var courseDTOs []dto.CourseResponseDTO
	for _, course := range courses {
		courseDTOs = append(courseDTOs, dto.CourseResponseDTO{
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
