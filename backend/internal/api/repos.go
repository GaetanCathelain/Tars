package api

import (
	"encoding/json"
	"net/http"
	"time"

	"tars/backend/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

type repoRow struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	GithubURL     string    `json:"github_url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *Handler) listRepos(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	rows, err := h.db.Query(r.Context(),
		`SELECT id, name, path, github_url, default_branch, created_at, updated_at
		 FROM repos WHERE owner_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	defer rows.Close()

	repos := []repoRow{}
	for rows.Next() {
		var rr repoRow
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.Path, &rr.GithubURL, &rr.DefaultBranch, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scan error", nil)
			return
		}
		repos = append(repos, rr)
	}

	writeJSON(w, http.StatusOK, map[string]any{"repos": repos})
}

func (h *Handler) createRepo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var body struct {
		Name      string `json:"name"`
		GithubURL string `json:"github_url"`
		Path      string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}
	if body.Name == "" || body.GithubURL == "" || body.Path == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name, github_url, and path are required", nil)
		return
	}

	id := "repo_" + ulid.Make().String()
	var rr repoRow
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO repos (id, owner_id, name, github_url, path)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, path, github_url, default_branch, created_at, updated_at`,
		id, userID, body.Name, body.GithubURL, body.Path,
	).Scan(&rr.ID, &rr.Name, &rr.Path, &rr.GithubURL, &rr.DefaultBranch, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "CONFLICT", "repository name already exists", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusCreated, rr)
}

func (h *Handler) getRepo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	var rr repoRow
	err := h.db.QueryRow(r.Context(),
		`SELECT id, name, path, github_url, default_branch, created_at, updated_at
		 FROM repos WHERE id = $1 AND owner_id = $2`,
		repoID, userID,
	).Scan(&rr.ID, &rr.Name, &rr.Path, &rr.GithubURL, &rr.DefaultBranch, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, rr)
}

func (h *Handler) updateRepo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	var body struct {
		Name          *string `json:"name"`
		DefaultBranch *string `json:"default_branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	var rr repoRow
	err := h.db.QueryRow(r.Context(),
		`UPDATE repos
		 SET name           = COALESCE($3, name),
		     default_branch = COALESCE($4, default_branch),
		     updated_at     = NOW()
		 WHERE id = $1 AND owner_id = $2
		 RETURNING id, name, path, github_url, default_branch, created_at, updated_at`,
		repoID, userID, body.Name, body.DefaultBranch,
	).Scan(&rr.ID, &rr.Name, &rr.Path, &rr.GithubURL, &rr.DefaultBranch, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
			return
		}
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "CONFLICT", "repository name already exists", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, rr)
}

func (h *Handler) deleteRepo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	tag, err := h.db.Exec(r.Context(),
		`DELETE FROM repos WHERE id = $1 AND owner_id = $2`,
		repoID, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isUniqueViolation checks for PostgreSQL unique constraint violation (error code 23505).
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return containsStr(err.Error(), "23505") || containsStr(err.Error(), "unique constraint")
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
