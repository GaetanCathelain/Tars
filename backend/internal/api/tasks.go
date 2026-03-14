package api

import (
	"encoding/json"
	"net/http"
	"time"

	"tars/backend/internal/auth"
	"tars/backend/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

type taskRow struct {
	ID          string    `json:"id"`
	RepoID      string    `json:"repo_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Priority    int       `json:"priority"`
	AgentID     *string   `json:"agent_id"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	q := `SELECT t.id, t.repo_id, t.title, t.description, t.status, t.priority,
	             t.agent_id, t.created_by, t.created_at, t.updated_at
	      FROM tasks t WHERE t.repo_id = $1`
	args := []any{repoID}
	argIdx := 2

	if status := r.URL.Query().Get("status"); status != "" {
		q += " AND t.status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		q += " AND t.agent_id = $" + itoa(argIdx)
		args = append(args, agentID)
		argIdx++
	}
	_ = argIdx

	q += " ORDER BY t.created_at DESC"

	rows, err := h.db.Query(r.Context(), q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	defer rows.Close()

	tasks := []taskRow{}
	for rows.Next() {
		var t taskRow
		if err := rows.Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status,
			&t.Priority, &t.AgentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scan error", nil)
			return
		}
		tasks = append(tasks, t)
	}

	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    *int   `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title is required", nil)
		return
	}

	priority := 3
	if body.Priority != nil {
		p := *body.Priority
		if p < 1 || p > 5 {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "priority must be between 1 and 5", nil)
			return
		}
		priority = p
	}

	id := "task_" + ulid.Make().String()
	var t taskRow
	err := h.db.QueryRow(r.Context(),
		`INSERT INTO tasks (id, repo_id, title, description, priority, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, repo_id, title, description, status, priority, agent_id, created_by, created_at, updated_at`,
		id, repoID, body.Title, nullStr(body.Description), priority, userID,
	).Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AgentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	// Broadcast task.created WebSocket event.
	h.broadcastTaskEvent("task.created", repoID, t)

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	taskID := chi.URLParam(r, "taskId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var t taskRow
	err := h.db.QueryRow(r.Context(),
		`SELECT id, repo_id, title, description, status, priority, agent_id, created_by, created_at, updated_at
		 FROM tasks WHERE id = $1 AND repo_id = $2`,
		taskID, repoID,
	).Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AgentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "task not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	taskID := chi.URLParam(r, "taskId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *int    `json:"priority"`
		AgentID     *string `json:"agent_id"` // null to unassign
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", nil)
		return
	}

	if body.Status != nil {
		validStatuses := map[string]bool{"pending": true, "in_progress": true, "done": true, "cancelled": true}
		if !validStatuses[*body.Status] {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid status value", nil)
			return
		}
	}
	if body.Priority != nil && (*body.Priority < 1 || *body.Priority > 5) {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "priority must be between 1 and 5", nil)
		return
	}

	var t taskRow
	err := h.db.QueryRow(r.Context(),
		`UPDATE tasks
		 SET title       = COALESCE($3, title),
		     description = COALESCE($4, description),
		     status      = COALESCE($5, status),
		     priority    = COALESCE($6, priority),
		     agent_id    = CASE WHEN $7::boolean THEN $8 ELSE agent_id END,
		     updated_at  = NOW()
		 WHERE id = $1 AND repo_id = $2
		 RETURNING id, repo_id, title, description, status, priority, agent_id, created_by, created_at, updated_at`,
		taskID, repoID,
		body.Title, body.Description, body.Status, body.Priority,
		body.AgentID != nil, body.AgentID, // CASE: only update agent_id if field was present
	).Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.AgentID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "task not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}

	// Broadcast task.updated WebSocket event.
	h.broadcastTaskEvent("task.updated", repoID, t)

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")
	taskID := chi.URLParam(r, "taskId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	tag, err := h.db.Exec(r.Context(),
		`DELETE FROM tasks WHERE id = $1 AND repo_id = $2`,
		taskID, repoID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "task not found", nil)
		return
	}

	// Broadcast task.deleted WebSocket event.
	if h.hub != nil {
		env := ws.Envelope{
			Type:    "task.deleted",
			Channel: "repo:" + repoID,
			Payload: mustMarshalAny(map[string]string{"task_id": taskID}),
		}
		h.hub.Broadcast("repo:"+repoID, env)
	}

	w.WriteHeader(http.StatusNoContent)
}

// repoOwnedBy checks that repoID exists and belongs to userID.
func (h *Handler) repoOwnedBy(r *http.Request, repoID, userID string) bool {
	var exists bool
	err := h.db.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM repos WHERE id = $1 AND owner_id = $2)`,
		repoID, userID,
	).Scan(&exists)
	return err == nil && exists
}

// broadcastTaskEvent sends a WS event for task mutations.
func (h *Handler) broadcastTaskEvent(eventType, repoID string, t taskRow) {
	if h.hub == nil {
		return
	}
	env := ws.Envelope{
		Type:    eventType,
		Channel: "repo:" + repoID,
		Payload: mustMarshalAny(map[string]any{"task": t}),
	}
	h.hub.Broadcast("repo:"+repoID, env)
}

func mustMarshalAny(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func itoa(i int) string {
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}
