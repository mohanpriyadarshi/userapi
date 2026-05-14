package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"userapi/db"
	"userapi/model"

	"golang.org/x/crypto/bcrypt"
)

// Handler contains dependencies for HTTP handlers, including the database.
type Handler struct {
	DB *db.DB
}

// ListUsers handles GET /users and returns a paginated list of users.
// It reads optional limit and offset query parameters and falls back to defaults.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 10 // default
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	users, err := h.DB.ListUsers(limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if users == nil {
		users = []model.User{}
	}
	respondJSON(w, http.StatusOK, users)
}

// CreateUser handles POST /users, validates the request body, hashes the password,
// and stores the new user record in the database.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "name, email and password are required")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.DB.CreateUser(req.Name, req.Email, string(hashed))
	if err != nil {
		respondError(w, http.StatusConflict, "email already exists")
		return
	}
	respondJSON(w, http.StatusCreated, user)
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	_, user, err := h.DB.GetUserByEmail(email)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// GetUser handles GET /users/{id} and returns the user for the provided ID.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.DB.GetUserByID(id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// UpdateUser handles PUT /users/{id}, validates input, and updates the user's profile.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "name and email are required")
		return
	}

	user, err := h.DB.UpdateUser(id, req.Name, req.Email)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id} and removes the specified user.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.DB.DeleteUser(id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// Login handles POST /login and validates the user's email and password.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	password, user, err := h.DB.GetUserByEmail(req.Email)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"message": "login successful",
		"user":    user,
	})
}

// respondJSON writes a JSON response with the given HTTP status.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response in JSON format.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
