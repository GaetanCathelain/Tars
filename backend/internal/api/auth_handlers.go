package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tars/backend/internal/auth"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Name      *string   `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// handleRegister handles POST /api/v1/auth/register.
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "valid email is required", nil)
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "password must be at least 8 characters", nil)
		return
	}

	// Check if email already in use.
	var existing string
	err := h.db.QueryRow(r.Context(), `SELECT id FROM users WHERE email = $1`, req.Email).Scan(&existing)
	if err == nil {
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "an account with that email already exists", nil)
		return
	}
	if err != pgx.ErrNoRows {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to hash password", nil)
		return
	}

	// Derive login from email local part.
	login := strings.SplitN(req.Email, "@", 2)[0]
	newID := "usr_" + ulid.Make().String()

	var u userResponse
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO users (id, login, name, email, password_hash)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, login, name, avatar_url, email, created_at`,
		newID, login, nullStr(req.Name), req.Email, hash,
	).Scan(&u.ID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &u.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create user", nil)
		return
	}

	if err := h.auth.CreateSession(w, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create session", nil)
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

// handleLogin handles POST /api/v1/auth/login.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "email and password are required", nil)
		return
	}

	var (
		u            userResponse
		passwordHash string
	)
	err := h.db.QueryRow(r.Context(),
		`SELECT id, login, name, avatar_url, email, password_hash, created_at
		 FROM users WHERE email = $1`,
		req.Email,
	).Scan(&u.ID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &passwordHash, &u.CreatedAt)
	if err == pgx.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	if passwordHash == "" {
		// GitHub-only user — no password set.
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
		return
	}

	if err := auth.CheckPassword(passwordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
		return
	}

	if err := h.auth.CreateSession(w, u.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create session", nil)
		return
	}

	writeJSON(w, http.StatusOK, u)
}
