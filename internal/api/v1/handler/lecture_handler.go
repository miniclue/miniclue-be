package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"app/internal/api/v1/dto"
	"app/internal/middleware"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
)

// LectureHandler handles flat lecture endpoints

type LectureHandler struct {
	lectureService     service.LectureService
	courseService      service.CourseService
	summaryService     service.SummaryService
	explanationService service.ExplanationService
	validate           *validator.Validate
}

// NewLectureHandler creates a new LectureHandler
func NewLectureHandler(
	lectureService     service.LectureService,
	courseService      service.CourseService,
	summaryService     service.SummaryService,
	explanationService service.ExplanationService,
	validate           *validator.Validate,
) *LectureHandler {
	return &LectureHandler{
		lectureService:     lectureService,
		courseService:      courseService,
		summaryService:     summaryService,
		explanationService: explanationService,
		validate:           validate,
	}
}

// RegisterRoutes mounts lecture routes under /lectures/{id}
func (h *LectureHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/lectures", authMw(http.HandlerFunc(h.listLectures)))
	mux.Handle("/lectures/", authMw(http.HandlerFunc(h.handleLecture)))
}

func (h *LectureHandler) handleLecture(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/lectures/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if strings.HasSuffix(path, "/summary") {
			h.getLectureSummary(w, r)
			return
		}
		if strings.HasSuffix(path, "/explanations") {
			h.listLectureExplanations(w, r)
			return
		}
		h.getLecture(w, r)
	case http.MethodPut:
		h.updateLecture(w, r)
	case http.MethodDelete:
		h.deleteLecture(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// getLecture godoc
// @Summary Get a lecture
// @Description Retrieves a lecture by its ID.
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Success 200 {object} dto.LectureResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to retrieve lecture"
// @Router /lectures/{lectureId} [get]
func (h *LectureHandler) getLecture(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimPrefix(r.URL.Path, "/lectures/")
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	// authorization: verify user owns course
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	resp := dto.LectureResponseDTO{
		LectureID:  lecture.ID,
		CourseID:   lecture.CourseID,
		Title:      lecture.Title,
		PdfURL:     lecture.PDFURL,
		Status:     lecture.Status,
		CreatedAt:  lecture.CreatedAt,
		UpdatedAt:  lecture.UpdatedAt,
		AccessedAt: lecture.AccessedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// updateLecture godoc
// @Summary Update a lecture
// @Description Updates lecture metadata.
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param lecture body dto.LectureUpdateDTO true "Lecture update data"
// @Success 200 {object} dto.LectureResponseDTO
// @Failure 400 {string} string "Invalid JSON payload"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to update lecture"
// @Router /lectures/{lectureId} [put]
func (h *LectureHandler) updateLecture(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimPrefix(r.URL.Path, "/lectures/")
	var req dto.LectureUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	if req.Title != nil {
		lecture.Title = *req.Title
	}
	if req.AccessedAt != nil {
		lecture.AccessedAt = *req.AccessedAt
	}
	if err := h.lectureService.UpdateLecture(r.Context(), lecture); err != nil {
		http.Error(w, "Failed to update lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.LectureResponseDTO{
		LectureID:  lecture.ID,
		CourseID:   lecture.CourseID,
		Title:      lecture.Title,
		PdfURL:     lecture.PDFURL,
		Status:     lecture.Status,
		CreatedAt:  lecture.CreatedAt,
		UpdatedAt:  lecture.UpdatedAt,
		AccessedAt: lecture.AccessedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// deleteLecture godoc
// @Summary Delete a lecture
// @Description Deletes a lecture and all its derived data.
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Success 204 {string} string "No Content"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to delete lecture"
// @Router /lectures/{lectureId} [delete]
func (h *LectureHandler) deleteLecture(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimPrefix(r.URL.Path, "/lectures/")
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	if err := h.lectureService.DeleteLecture(r.Context(), lectureID); err != nil {
		http.Error(w, "Failed to delete lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listLectures godoc
// @Summary List lectures
// @Description Retrieves lectures filtered by course_id with pagination
// @Tags lectures
// @Produce json
// @Param course_id query string true "Course ID"
// @Param limit query int false "Limit number of lectures"
// @Param offset query int false "Pagination offset"
// @Success 200 {array} dto.LectureResponseDTO
// @Failure 400 {string} string "Missing or invalid course_id"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 500 {string} string "Failed to retrieve lectures"
// @Router /lectures [get]
func (h *LectureHandler) listLectures(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	q := r.URL.Query()
	courseID := q.Get("course_id")
	if courseID == "" {
		http.Error(w, "Missing course_id", http.StatusBadRequest)
		return
	}
	// authorization: verify user owns this course
	course, err := h.courseService.GetCourseByID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Failed to retrieve course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if course == nil || course.UserID != userID {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	limit := 10
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	lectures, err := h.lectureService.GetLecturesByCourseID(r.Context(), courseID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve lectures: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var resp []dto.LectureResponseDTO
	for _, lec := range lectures {
		resp = append(resp, dto.LectureResponseDTO{
			LectureID:  lec.ID,
			CourseID:   lec.CourseID,
			Title:      lec.Title,
			PdfURL:     lec.PDFURL,
			Status:     lec.Status,
			CreatedAt:  lec.CreatedAt,
			UpdatedAt:  lec.UpdatedAt,
			AccessedAt: lec.AccessedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getLectureSummary godoc
// @Summary Get lecture summary
// @Description Retrieves a lecture's summary by its ID.
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Success 200 {object} dto.LectureSummaryResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to retrieve summary"
// @Router /lectures/{lectureId}/summary [get]
func (h *LectureHandler) getLectureSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/summary")
	// verify lecture exists and ownership
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	// fetch summary
	summary, err := h.summaryService.GetSummaryByLectureID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve summary: "+err.Error(), http.StatusInternalServerError)
		return
	}
	content := ""
	if summary != nil {
		content = summary.Content
	}
	resp := dto.LectureSummaryResponseDTO{LectureID: lectureID, Content: content}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// listLectureExplanations godoc
// @Summary List lecture explanations
// @Description Retrieves explanations for a lecture with pagination
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param limit query int false "Limit number of results"
// @Param offset query int false "Pagination offset"
// @Success 200 {array} dto.LectureExplanationResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to retrieve explanations"
// @Router /lectures/{lectureId}/explanations [get]
func (h *LectureHandler) listLectureExplanations(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	// extract lectureId
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/explanations")
	// verify lecture exists
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	// authorization: verify user owns this course
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	// parse query params
	q := r.URL.Query()
	limit := 10
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	explanations, err := h.explanationService.GetExplanationsByLectureID(r.Context(), lectureID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to retrieve explanations: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var resp []dto.LectureExplanationResponseDTO
	for _, e := range explanations {
		resp = append(resp, dto.LectureExplanationResponseDTO{
			ID:          e.ID,
			LectureID:   e.LectureID,
			SlideNumber: e.SlideNumber,
			Content:     e.Content,
			CreatedAt:   e.CreatedAt,
			UpdatedAt:   e.UpdatedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
