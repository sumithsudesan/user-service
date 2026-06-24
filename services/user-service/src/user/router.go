package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler struct that holds a reference to the Service.
type Handler struct {
	service *Service
}

// NewHandler creates a new Handler with the provided Service.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Request and Response structs for API endpoints
type createRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

// Request struct for update endpoint
type updateRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Status  string `json:"status"`
	Version int    `json:"version"`
}

// NewRouter creates a new router with the provided Handler.
func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Post("/users", h.CreateUser)
	r.Get("/users", h.ListUsers)
	r.Get("/users/{id}", h.GetUser)
	r.Put("/users/{id}", h.UpdateUser)
	r.Delete("/users/{id}", h.DeleteUser)

	return r
}

// CreateUser handles the creation of a new user.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	user, err := h.service.Create(CreateInput{
		Name:   req.Name,
		Email:  req.Email,
		Status: req.Status,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrNameRequired), errors.Is(err, ErrEmailRequired):
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// GetUser handles fetching a user by ID.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.service.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ListUsers handles fetching all users.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

// UpdateUser handles updating a user by ID.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequest)
		return
	}

	user, err := h.service.Update(id, UpdateInput{
		Name:    req.Name,
		Email:   req.Email,
		Status:  req.Status,
		Version: req.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, err)
		case errors.Is(err, ErrVersionMismatch):
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// DeleteUser handles deleting a user by ID.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.service.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(w, http.StatusNotFound, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions for writing JSON responses and errors
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes an error response in JSON format.
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}
