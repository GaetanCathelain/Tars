package api

import (
	"context"
	"net/http"
	"time"

	"tars/backend/internal/auth"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

// handleGitHubLogin initiates the GitHub OAuth flow.
func (h *Handler) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	h.github.HandleLogin(w, r)
}

// handleGitHubCallback handles the OAuth callback from GitHub.
func (h *Handler) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	h.github.HandleCallback(w, r, h.upsertUser)
}

// handleMe returns the currently authenticated user.
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	row := h.db.QueryRow(r.Context(),
		`SELECT id, github_id, login, name, avatar_url, email, created_at FROM users WHERE id = $1`,
		userID,
	)

	var u struct {
		ID        string    `json:"id"`
		GitHubID  *int64    `json:"github_id"`
		Login     string    `json:"login"`
		Name      *string   `json:"name"`
		AvatarURL *string   `json:"avatar_url"`
		Email     *string   `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := row.Scan(&u.ID, &u.GitHubID, &u.Login, &u.Name, &u.AvatarURL, &u.Email, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// handleLogout clears the session cookie.
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	h.auth.ClearSession(w)
	w.WriteHeader(http.StatusNoContent)
}

// upsertUser inserts or updates a user from GitHub OAuth data and returns the TARS user ID.
func (h *Handler) upsertUser(ctx context.Context, u auth.GitHubUser) (string, error) {
	// Check if user already exists by github_id.
	var existingID string
	err := h.db.QueryRow(ctx,
		`SELECT id FROM users WHERE github_id = $1`,
		u.ID,
	).Scan(&existingID)

	if err == nil {
		// User exists — update their profile.
		_, err = h.db.Exec(ctx,
			`UPDATE users SET login=$1, name=$2, avatar_url=$3, email=$4, updated_at=NOW() WHERE id=$5`,
			u.Login, nullStr(u.Name), nullStr(u.AvatarURL), nullStr(u.Email), existingID,
		)
		return existingID, err
	}

	if err != pgx.ErrNoRows {
		return "", err
	}

	// New user — generate ULID ID and insert.
	newID := "usr_" + ulid.Make().String()
	_, err = h.db.Exec(ctx,
		`INSERT INTO users (id, github_id, login, name, avatar_url, email)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		newID, u.ID, u.Login, nullStr(u.Name), nullStr(u.AvatarURL), nullStr(u.Email),
	)
	if err != nil {
		return "", err
	}

	return newID, nil
}

// nullStr converts an empty string to nil for nullable TEXT columns.
func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
