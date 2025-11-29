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
	"github.com/rs/zerolog"
)

// LectureHandler handles flat lecture endpoints

type LectureHandler struct {
	lectureService     service.LectureService
	courseService      service.CourseService
	summaryService     service.SummaryService
	explanationService service.ExplanationService
	noteService        service.NoteService
	chatHandler        *ChatHandler
	validate           *validator.Validate
	s3BaseURL          string
	s3Bucket           string
	logger             zerolog.Logger
}

// NewLectureHandler creates a new LectureHandler
func NewLectureHandler(
	lectureService service.LectureService,
	courseService service.CourseService,
	summaryService service.SummaryService,
	explanationService service.ExplanationService,
	noteService service.NoteService,
	chatHandler *ChatHandler,
	validate *validator.Validate,
	s3BaseURL string,
	s3Bucket string,
	logger zerolog.Logger,
) *LectureHandler {
	return &LectureHandler{
		lectureService:     lectureService,
		courseService:      courseService,
		summaryService:     summaryService,
		explanationService: explanationService,
		noteService:        noteService,
		chatHandler:        chatHandler,
		validate:           validate,
		s3BaseURL:          s3BaseURL,
		s3Bucket:           s3Bucket,
		logger:             logger,
	}
}

// RegisterRoutes mounts lecture routes under /lectures/{id} with auth middleware
func (h *LectureHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/lectures", authMw(http.HandlerFunc(h.handleLectures)))
	mux.Handle("/lectures/", authMw(http.HandlerFunc(h.handleLecture)))
}

func (h *LectureHandler) handleLecture(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/lectures/") {
		http.NotFound(w, r)
		return
	}
	// Delegate chat routes to ChatHandler
	if strings.Contains(path, "/chats") {
		if h.chatHandler != nil {
			h.chatHandler.handleChatRoutes(w, r)
			return
		}
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
		if strings.HasSuffix(path, "/url") {
			h.getSignedURL(w, r)
			return
		}
		h.getLecture(w, r)
	case http.MethodPatch:
		if strings.HasSuffix(path, "/note") {
			h.updateLectureNote(w, r)
			return
		}
		h.updateLecture(w, r)
	case http.MethodPost:
		if strings.HasSuffix(path, "/note") {
			h.createLectureNote(w, r)
			return
		}
		if strings.HasSuffix(path, "/upload-complete") {
			h.completeUpload(w, r)
			return
		}
		if strings.HasSuffix(path, "/batch-upload-url") {
			h.getBatchUploadURL(w, r)
			return
		}
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
		LectureID:   lecture.ID,
		CourseID:    lecture.CourseID,
		Title:       lecture.Title,
		StoragePath: lecture.StoragePath,
		Status:      lecture.Status,
		CreatedAt:   lecture.CreatedAt,
		UpdatedAt:   lecture.UpdatedAt,
		AccessedAt:  lecture.AccessedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// updateLecture godoc
// @Summary Update a lecture
// @Description Updates lecture metadata including title, accessed_at, and course_id.
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param lecture body dto.LectureUpdateDTO true "Lecture update data"
// @Success 200 {object} dto.LectureResponseDTO
// @Failure 400 {string} string "Invalid JSON payload, title cannot be empty, or course_id cannot be empty"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found or course not found"
// @Failure 500 {string} string "Failed to update lecture"
// @Router /lectures/{lectureId} [patch]
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
	if req.Title != nil && strings.TrimSpace(*req.Title) == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}
	if req.CourseID != nil && strings.TrimSpace(*req.CourseID) == "" {
		http.Error(w, "Course ID cannot be empty", http.StatusBadRequest)
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
	// Verify user owns the current course
	currentCourse, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || currentCourse == nil || currentCourse.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}
	// If course_id is being updated, verify the new course exists and belongs to the user
	if req.CourseID != nil {
		newCourse, err := h.courseService.GetCourseByID(r.Context(), *req.CourseID)
		if err != nil {
			http.Error(w, "Failed to retrieve new course: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if newCourse == nil || newCourse.UserID != userID {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		lecture.CourseID = *req.CourseID
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
		LectureID:   lecture.ID,
		CourseID:    lecture.CourseID,
		Title:       lecture.Title,
		StoragePath: lecture.StoragePath,
		Status:      lecture.Status,
		CreatedAt:   lecture.CreatedAt,
		UpdatedAt:   lecture.UpdatedAt,
		AccessedAt:  lecture.AccessedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// deleteLecture godoc
// @Summary Delete a lecture
// @Description Deletes a lecture and all its derived database records, removes its PDF from storage, and clears related pending jobs from ingestion, embedding, explanation, and summary queues.
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

// handleLectures routes GET for listing lectures
func (h *LectureHandler) handleLectures(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listLectures(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
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
			LectureID:   lec.ID,
			CourseID:    lec.CourseID,
			Title:       lec.Title,
			StoragePath: lec.StoragePath,
			Status:      lec.Status,
			CreatedAt:   lec.CreatedAt,
			UpdatedAt:   lec.UpdatedAt,
			AccessedAt:  lec.AccessedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// listLectureExplanations godoc
// @Summary List lecture explanations
// @Description Retrieves explanations for a lecture with pagination
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param limit query int false "Limit number of results (if omitted, returns all explanations)"
// @Param offset query int false "Pagination offset (default 0)"
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
	// default to a high limit to fetch all if not specified
	limit := 1000
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// updateLectureNote godoc
// @Summary Update a lecture note
// @Description Updates the content of a note for a lecture.
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param note body dto.LectureNoteUpdateDTO true "Note update data"
// @Success 200 {object} dto.LectureNoteResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to update note"
// @Router /lectures/{lectureId}/note [patch]
func (h *LectureHandler) updateLectureNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/notes")
	// verify lecture exists and belongs to user
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
	// parse update payload
	var req dto.LectureNoteUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	// update and persist via lectureId-only
	updated, err := h.noteService.UpdateNoteByLectureID(r.Context(), lectureID, req.Content)
	if err != nil {
		http.Error(w, "Failed to update note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// respond
	resp := dto.LectureNoteResponseDTO{
		ID:        updated.ID,
		LectureID: updated.LectureID,
		Content:   updated.Content,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// createLectureNote godoc
// @Summary Create a lecture note
// @Description Creates a new note for a lecture.
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param note body dto.LectureNoteCreateDTO true "Note create data"
// @Success 201 {object} dto.LectureNoteResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to create note"
// @Router /lectures/{lectureId}/note [post]
func (h *LectureHandler) createLectureNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/notes")
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
	// Prevent duplicate note for this lecture
	existing, err := h.noteService.GetNoteByLectureID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to check existing note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, "Note already exists for this lecture", http.StatusConflict)
		return
	}
	var req dto.LectureNoteCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	created, err := h.noteService.CreateNoteByLectureID(r.Context(), userID, lectureID, req.Content)
	if err != nil {
		http.Error(w, "Failed to create note: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.LectureNoteResponseDTO{
		ID:        created.ID,
		LectureID: created.LectureID,
		Content:   created.Content,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// getSignedURL godoc
// @Summary Get signed URL for lecture PDF
// @Description Generates a signed URL for downloading the lecture PDF.
// @Tags lectures
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Success 200 {object} dto.SignedURLResponseDTO
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found"
// @Failure 500 {string} string "Failed to generate signed URL"
// @Router /lectures/{lectureId}/url [get]
func (h *LectureHandler) getSignedURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/url")
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
	url, err := h.lectureService.GetPresignedURL(r.Context(), lecture.StoragePath)
	if err != nil {
		http.Error(w, "Failed to generate signed URL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.SignedURLResponseDTO{URL: url}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// getBatchUploadURL godoc
// @Summary Get upload URLs for lectures
// @Description Initiates lecture uploads by creating lecture records and returning presigned URLs for direct S3 upload. Works for both single and multiple files.
// @Tags lectures
// @Accept json
// @Produce json
// @Param request body dto.LectureUploadURLRequestDTO true "Upload URL request"
// @Success 201 {object} dto.LectureBatchUploadURLResponseDTO
// @Failure 400 {string} string "Invalid JSON payload or validation failed"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 403 {string} string "Upload limit exceeded"
// @Failure 404 {string} string "Course not found or access denied"
// @Failure 500 {string} string "Failed to create upload URLs"
// @Router /lectures/batch-upload-url [post]
func (h *LectureHandler) getBatchUploadURL(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req dto.LectureUploadURLRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validate.Struct(&req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify course exists and user has access
	course, err := h.courseService.GetCourseByID(r.Context(), req.CourseID)
	if err != nil {
		http.Error(w, "Failed to retrieve course: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if course == nil || course.UserID != userID {
		http.Error(w, "Course not found or access denied", http.StatusNotFound)
		return
	}

	// Initiate batch upload
	lectures, presignedURLs, err := h.lectureService.InitiateBatchUpload(r.Context(), req.CourseID, userID, req.Filenames)
	if err != nil {
		http.Error(w, "Failed to create batch upload URLs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record upload events for all files
	for i := 0; i < len(req.Filenames); i++ {
		err := h.lectureService.RecordUploadEvent(r.Context(), userID, lectures[i].ID)
		if err != nil {
			// If recording fails, clean up created lectures
			for _, lecture := range lectures {
				_ = h.lectureService.DeleteLecture(r.Context(), lecture.ID)
			}
			http.Error(w, "Failed to record upload events: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Build response
	var uploads []dto.LectureUploadURLResponseDTO
	for i, lecture := range lectures {
		uploads = append(uploads, dto.LectureUploadURLResponseDTO{
			LectureID: lecture.ID,
			UploadURL: presignedURLs[i],
		})
	}

	resp := dto.LectureBatchUploadURLResponseDTO{
		Uploads: uploads,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}

// completeUpload godoc
// @Summary Complete lecture upload
// @Description Finalizes a lecture upload by verifying the file exists in storage and triggering processing.
// @Tags lectures
// @Accept json
// @Produce json
// @Param lectureId path string true "Lecture ID"
// @Param request body dto.LectureUploadCompleteRequestDTO true "Upload complete request"
// @Success 200 {object} dto.LectureUploadCompleteResponseDTO
// @Failure 400 {string} string "Invalid JSON payload"
// @Failure 401 {string} string "Unauthorized: User ID not found in context"
// @Failure 404 {string} string "Lecture not found or access denied"
// @Failure 500 {string} string "Failed to complete upload"
// @Router /lectures/{lectureId}/upload-complete [post]
func (h *LectureHandler) completeUpload(w http.ResponseWriter, r *http.Request) {
	// Authenticate
	userID, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized: User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Extract lecture ID from path
	lectureID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/lectures/"), "/upload-complete")

	// Verify lecture exists and user has access
	lecture, err := h.lectureService.GetLectureByID(r.Context(), lectureID)
	if err != nil {
		http.Error(w, "Failed to retrieve lecture: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if lecture == nil {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}

	// Verify user owns the course
	course, err := h.courseService.GetCourseByID(r.Context(), lecture.CourseID)
	if err != nil || course == nil || course.UserID != userID {
		http.Error(w, "Lecture not found", http.StatusNotFound)
		return
	}

	// Complete the upload
	updatedLecture, err := h.lectureService.CompleteUpload(r.Context(), lectureID, userID)
	if err != nil {
		http.Error(w, "Failed to complete upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.LectureUploadCompleteResponseDTO{
		LectureID: updatedLecture.ID,
		CourseID:  updatedLecture.CourseID,
		Status:    updatedLecture.Status,
		Message:   "Upload completed successfully. Processing has been initiated.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode response")
	}
}
