package handler

import (
	"encoding/json"
	"net/http"
	"strings"

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
	json.NewEncoder(w).Encode(resp)
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
	json.NewEncoder(w).Encode(resp)
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
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Course not found"
// @Failure 500 {string} string "Failed to update course"
// @Router /courses/{courseId} [put]
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
	json.NewEncoder(w).Encode(resp)
}

func (h *CourseHandler) handleCourse(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/courses/") {
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
	case http.MethodPut:
		h.updateCourse(w, r)
	default:
		http.NotFound(w, r)
	}
}
