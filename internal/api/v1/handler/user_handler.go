package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"app/internal/api/v1/dto"
	"app/internal/middleware"
	// "app/internal/model"
	"app/internal/service"

	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc service.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{svc: svc, validate: v}
}

// RegisterRoutes mounts v1 user routes
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	mux.Handle("/users/me", authMw(http.HandlerFunc(h.handleUsers)))
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
	// var req dto.UserCreateDTO
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, "invalid JSON payload", http.StatusBadRequest)
	// 	return
	// }
	// if err := h.validate.Struct(&req); err != nil {
	// 	http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// // Map DTO → domain model
	// user := &model.User{
	// 	Email:    req.Email,
	// 	Name:     req.Name,
	// 	Password: req.Password, // service will hash
	// }

	// created, err := h.svc.Register(r.Context(), user)
	// if err != nil {
	// 	switch {
	// 	case errors.Is(err, service.ErrEmailAlreadyRegistered):
	// 		http.Error(w, err.Error(), http.StatusConflict)
	// 	default:
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	}
	// 	return
	// }

	// // Map domain → response DTO
	// resp := dto.UserResponseDTO{
	// 	ID:        created.ID,
	// 	Email:     created.Email,
	// 	Name:      created.Name,
	// 	CreatedAt: created.CreatedAt,
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(middleware.UserContextKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	user, err := h.svc.GetUser(r.Context(), userId)
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
