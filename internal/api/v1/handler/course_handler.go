package handler

import (
	"encoding/json"
	"net/http"

	"app/internal/api/v1/dto"
	"app/internal/model"
	"app/internal/service"
	"app/internal/middleware"

	"github.com/go-playground/validator/v10"
)

// CourseHandler handles course-related endpoints
type CourseHandler struct {
	courseService service.CourseService
	validate      *validator.Validate
}

// NewCourseHandler creates a new CourseHandler
func NewCourseHandler(courseService service.CourseService, validate *validator.Validate) *CourseHandler {
	return &CourseHandler{courseService: courseService, validate: validate}
}

// RegisterRoutes mounts course routes
func (h *CourseHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/courses", authMw(http.HandlerFunc(h.handleCourses)))
}

// handleCourses handles POST /courses
func (h *CourseHandler) handleCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/courses" {
		http.NotFound(w, r)
		return
	}
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	var req dto.CourseCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Build model
	description := ""
	if req.Description != nil {
		description = *req.Description
	}
	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}
	course := &model.Course{
		UserID:      userID,
		Title:       req.Title,
		Description: description,
		IsDefault:   isDefault,
	}
	created, err := h.courseService.CreateCourse(r.Context(), course)
	if err != nil {
		http.Error(w, "Failed to create course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.CourseResponseDTO{
		CourseID:    created.CourseID,
		UserID:      created.UserID,
		Title:       created.Title,
		Description: created.Description,
		IsDefault:   created.IsDefault,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
