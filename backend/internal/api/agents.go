package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"tars/backend/internal/agent"
	"tars/backend/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

type agentRow struct {
	ID           string     `json:"id"`
	RepoID       string     `json:"repo_id"`
	TaskID       *string    `json:"task_id"`
	Name         string     `json:"name"`
	Persona      *string    `json:"persona"`
	Model        string     `json:"model"`
	Status       string     `json:"status"`
	WorktreePath *string    `json:"worktree_path"`
	Branch       *string    `json:"branch"`
	PID          *int       `json:"pid"`
	StartedAt    *time.Time `json:"started_at"`
	StoppedAt    *time.Time `json:"stopped_at"`
}

const agentScanCols = `id, repo_id, task_id, name, persona, model, status, worktree_path, branch, pid, started_at, stopped_at`

func scanAgent(row interface{ Scan(...any) error }) (agentRow, error) {
	var a agentRow
	err := row.Scan(&a.ID, &a.RepoID, &a.TaskID, &a.Name, &a.Persona, &a.Model,
		&a.Status, &a.WorktreePath, &a.Branch, &a.PID, &a.StartedAt, &a.StoppedAt)
	return a, err
}

func (h *Handler) listAgents(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	q := `SELECT ` + agentScanCols + ` FROM agents WHERE repo_id = $1`
	args := []any{repoID}
	if status := r.URL.Query().Get("status"); status != "" {
		q += " AND status = $2"
		args = append(args, status)
	}
	q += " ORDER BY started_at DESC NULLS LAST"

	rows, err := h.db.Query(r.Context(), q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	defer rows.Close()

	agents := []agentRow{}
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scan error", nil)
			return
		}
		agents = append(agents, a)
	}

	writeJSON(w, http.StatusOK, map[string]any{"agents": agents})
}

func (h *Handler) spawnAgent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var body struct {
		TaskID       *string `json:"task_id"`
		Name         string  `json:"name"`
		Persona      *string `json:"persona"`
		Model        string  `json:"model"`
		SystemPrompt *string `json:"system_prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required", nil)
		return
	}
	if body.Model == "" {
		body.Model = "claude-opus-4-5"
	}

	// Get repo path for worktree creation.
	var repoPath, defaultBranch string
	err := h.db.QueryRow(r.Context(),
		`SELECT path, default_branch FROM repos WHERE id = $1`, repoID,
	).Scan(&repoPath, &defaultBranch)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	// Insert agent record with status "starting".
	agentID := "agent_" + ulid.Make().String()
	var a agentRow
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO agents (id, repo_id, task_id, name, persona, model, system_prompt, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'starting')
		 RETURNING `+agentScanCols,
		agentID, repoID, body.TaskID, body.Name, body.Persona, body.Model, body.SystemPrompt,
	).Scan(&a.ID, &a.RepoID, &a.TaskID, &a.Name, &a.Persona, &a.Model,
		&a.Status, &a.WorktreePath, &a.Branch, &a.PID, &a.StartedAt, &a.StoppedAt)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "CONFLICT", "agent name already exists in this repo", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	// Create git worktree.
	var worktreePath, branch string
	if h.worktree != nil {
		wt, err := h.worktree.Create(r.Context(), repoPath, agentID)
		if err != nil {
			// Rollback DB record.
			h.db.Exec(r.Context(), `DELETE FROM agents WHERE id = $1`, agentID)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create worktree: "+err.Error(), nil)
			return
		}
		worktreePath = wt.Path
		branch = wt.Branch

		h.db.Exec(r.Context(),
			`UPDATE agents SET worktree_path=$2, branch=$3, updated_at=NOW() WHERE id=$1`,
			agentID, worktreePath, branch,
		)
		wtPath := worktreePath
		a.WorktreePath = &wtPath
		a.Branch = &branch
	}

	// Spawn the agent process.
	if h.agents != nil {
		agentRecord := agent.Agent{
			ID:           agentID,
			RepoID:       repoID,
			TaskID:       body.TaskID,
			Name:         body.Name,
			Model:        body.Model,
			SystemPrompt: body.SystemPrompt,
			WorktreePath: a.WorktreePath,
		}
		_, err = h.agents.Spawn(r.Context(), agentRecord)
		if err != nil {
			// Cleanup worktree and DB.
			if h.worktree != nil {
				h.worktree.Remove(r.Context(), repoPath, agentID)
			}
			h.db.Exec(r.Context(), `DELETE FROM agents WHERE id = $1`, agentID)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to spawn agent: "+err.Error(), nil)
			return
		}
	}

	// Refresh the agent record from DB to get final state.
	h.db.QueryRow(r.Context(),
		`SELECT `+agentScanCols+` FROM agents WHERE id = $1`, agentID,
	).Scan(&a.ID, &a.RepoID, &a.TaskID, &a.Name, &a.Persona, &a.Model,
		&a.Status, &a.WorktreePath, &a.Branch, &a.PID, &a.StartedAt, &a.StoppedAt)

	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	a, err := scanAgent(h.db.QueryRow(r.Context(),
		`SELECT `+agentScanCols+` FROM agents WHERE id = $1 AND repo_id = $2`,
		agentID, repoID,
	))
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) stopAgent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	if h.agents != nil {
		if err := h.agents.Stop(r.Context(), agentID, repoID); err != nil {
			writeError(w, http.StatusBadRequest, "AGENT_NOT_RUNNING", err.Error(), nil)
			return
		}
	}

	a, err := scanAgent(h.db.QueryRow(r.Context(),
		`SELECT `+agentScanCols+` FROM agents WHERE id = $1 AND repo_id = $2`,
		agentID, repoID,
	))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) sendAgentInput(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "text is required", nil)
		return
	}

	if h.agents != nil {
		if err := h.agents.SendInput(agentID, body.Text); err != nil {
			writeError(w, http.StatusBadRequest, "AGENT_NOT_RUNNING", "agent is not running", nil)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getAgentLogs(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	limit := 500
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > 5000 {
				v = 5000
			}
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var lines []agent.LogLine
	var total int64
	if h.agents != nil {
		lines, total = h.agents.Logs(agentID, limit, offset)
	}
	if lines == nil {
		lines = []agent.LogLine{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"agent_id": agentID,
		"lines":    lines,
		"total":    total,
	})
}

func (h *Handler) mergeAgent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var body struct {
		TargetBranch  string `json:"target_branch"`
		Strategy      string `json:"strategy"`
		CommitMessage string `json:"commit_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}
	if body.TargetBranch == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "target_branch is required", nil)
		return
	}
	if body.Strategy == "" {
		body.Strategy = "squash"
	}

	// Get agent branch from DB.
	var agentBranch string
	var repoPath string
	err := h.db.QueryRow(r.Context(),
		`SELECT a.branch, r.path FROM agents a JOIN repos r ON r.id = a.repo_id WHERE a.id = $1 AND a.repo_id = $2`,
		agentID, repoID,
	).Scan(&agentBranch, &repoPath)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	// Perform merge using git.
	commitSHA, err := gitMerge(r.Context(), repoPath, agentBranch, body.TargetBranch, body.Strategy, body.CommitMessage)
	if err != nil {
		writeError(w, http.StatusConflict, "MERGE_CONFLICT", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"merged":       true,
		"target_branch": body.TargetBranch,
		"agent_branch": agentBranch,
		"commit_sha":   commitSHA,
	})
}

func (h *Handler) getAgentDiff(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	agentID := chi.URLParam(r, "agentId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "unified"
	}

	// Get repo path and default branch.
	var repoPath, defaultBranch, agentBranch string
	err := h.db.QueryRow(r.Context(),
		`SELECT r.path, r.default_branch, a.branch
		 FROM agents a JOIN repos r ON r.id = a.repo_id
		 WHERE a.id = $1 AND a.repo_id = $2`,
		agentID, repoID,
	).Scan(&repoPath, &defaultBranch, &agentBranch)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	baseRef := r.URL.Query().Get("base")
	if baseRef == "" {
		baseRef = defaultBranch
	}

	result, err := gitDiff(repoPath, baseRef, agentBranch, format, agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "diff error: "+err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
