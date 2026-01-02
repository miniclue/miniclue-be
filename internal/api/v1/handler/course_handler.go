package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"app/internal/api/v1/dto"
	"app/internal/middleware"
	"app/internal/model"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

// CourseHandler handles course-related endpoints
type CourseHandler struct {
	courseService service.CourseService
	validate      *validator.Validate
	logger        zerolog.Logger
}

// NewCourseHandler creates a new CourseHandler
func NewCourseHandler(courseService service.CourseService, validate *validator.Validate, logger zerolog.Logger) *CourseHandler {
	return &CourseHandler{
		courseService: courseService,
		validate:      validate,
		logger:        logger,
	}
}

// RegisterRoutes mounts course routes
func (h *CourseHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/courses", authMw(http.HandlerFunc(h.createCourse)))
	mux.Handle("/courses/", authMw(http.HandlerFunc(h.handleCourse)))
}

// createCourse godoc
// @Summary Create a new course
// @Description Creates a new course associated with the authenticated user.
// @Tags courses
// @Accept json
// @Produce json
// @Param course body dto.CourseCreateDTO true "Course creation request"
// @Success 201 {object} dto.CourseResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 500 {string} string "Failed to create course"
// @Router /courses [post]
func (h *CourseHandler) createCourse(w http.ResponseWriter, r *http.Request) {
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

func (h *CourseHandler) handleCourse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/courses/") {
		http.NotFound(w, r)
		return
	}
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getCourse(w, r)
	case http.MethodPatch:
		h.updateCourse(w, r)
	case http.MethodDelete:
		h.deleteCourse(w, r)
	default:
		http.NotFound(w, r)
	}
}

// getCourseByID godoc
// @Summary Get a course
// @Description Retrieves a course by its ID.
// @Tags courses
// @Produce json
// @Param courseId path string true "Course ID"
// @Success 200 {object} dto.CourseResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Course not found"
// @Failure 500 {string} string "Failed to retrieve course"
// @Router /courses/{courseId} [get]
func (h *CourseHandler) getCourse(w http.ResponseWriter, r *http.Request) {
	courseID := strings.TrimPrefix(r.URL.Path, "/courses/")
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Failed to retrieve course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if course == nil || course.UserID != userID {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	resp := dto.CourseResponseDTO{
		CourseID:    course.CourseID,
		UserID:      course.UserID,
		Title:       course.Title,
		Description: course.Description,
		IsDefault:   course.IsDefault,
		CreatedAt:   course.CreatedAt,
		UpdatedAt:   course.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// updateCourse godoc
// @Summary Update a course
// @Description Updates an existing course by its ID.
// @Tags courses
// @Accept json
// @Produce json
// @Param courseId path string true "Course ID"
// @Param course body dto.CourseUpdateDTO true "Course update request"
// @Success 200 {object} dto.CourseResponseDTO
// @Failure 400 {string} string "Invalid JSON payload, validation failed, or title cannot be empty"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Course not found"
// @Failure 500 {string} string "Failed to update course"
// @Router /courses/{courseId} [patch]
func (h *CourseHandler) updateCourse(w http.ResponseWriter, r *http.Request) {
	courseID := strings.TrimPrefix(r.URL.Path, "/courses/")
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	var req dto.CourseUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Failed to retrieve course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if course == nil || course.UserID != userID {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if req.Title != nil {
		course.Title = *req.Title
	}
	if req.Description != nil {
		course.Description = *req.Description
	}
	if req.IsDefault != nil {
		course.IsDefault = *req.IsDefault
	}
	updated, err := h.courseService.UpdateCourse(r.Context(), course)
	if err != nil {
		http.Error(w, "Failed to update course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.CourseResponseDTO{
		CourseID:    updated.CourseID,
		UserID:      updated.UserID,
		Title:       updated.Title,
		Description: updated.Description,
		IsDefault:   updated.IsDefault,
		CreatedAt:   updated.CreatedAt,
		UpdatedAt:   updated.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// deleteCourse godoc
// @Summary Delete a course
// @Description Deletes a course and all its lectures, removes associated PDFs from storage, clears any pending jobs in ingestion and embedding queues, and deletes related database records.
// @Tags courses
// @Produce json
// @Param courseId path string true "Course ID"
// @Success 204 {string} string "No Content"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Course not found"
// @Failure 500 {string} string "Failed to delete course"
// @Router /courses/{courseId} [delete]
func (h *CourseHandler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	courseID := strings.TrimPrefix(r.URL.Path, "/courses/")
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Failed to retrieve course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if course == nil || course.UserID != userID {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if err := h.courseService.DeleteCourse(r.Context(), courseID); err != nil {
		http.Error(w, "Failed to delete course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
