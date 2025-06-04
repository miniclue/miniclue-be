package handler

import (
	"app/internal/model"
	"app/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(us service.UserService, validate *validator.Validate) *UserHandler {
	return &UserHandler{userService: us, validate: validate}
}

// RegisterRoutes mounts user-related routes onto an http.ServeMux
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/users/", h.handleUsers)
	mux.HandleFunc("/api/users/{userID}", h.handleUserByID) // Path-based parameters for ServeMux
}

func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/users/" { // Ensure exact match for POST to /api/users/
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		h.CreateUser(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) handleUserByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.GetUser(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CreateUser handles POST /api/users/
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var dto model.UserCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(&dto); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	userResp, err := h.userService.Register(ctx, &dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResp)
}

// GetUser handles GET /api/users/{userID}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract userID from path: /api/users/{userID}
	// For ServeMux, path parameters are handled differently. We get them from r.URL.Path.
	// The pattern registered is "/api/users/{userID}" which will match /api/users/anyValueHere
	// We need to parse "anyValueHere" part.
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/") // e.g. ["api", "users", "123"]

	if len(parts) < 3 {
		http.Error(w, "invalid user ID in path", http.StatusBadRequest)
		return
	}
	idParam := parts[2] // Assumes path is /api/users/{id}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		http.Error(w, "invalid user ID format: "+idParam, http.StatusBadRequest)
		return
	}
	userResp, err := h.userService.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if userResp == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResp)
}
