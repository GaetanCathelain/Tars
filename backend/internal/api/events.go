package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"tars/backend/internal/auth"

	"github.com/go-chi/chi/v5"
)

type eventRow struct {
	ID        string         `json:"id"`
	RepoID    string         `json:"repo_id"`
	Type      string         `json:"type"`
	ActorType string         `json:"actor_type"`
	ActorID   *string        `json:"actor_id"`
	AgentID   *string        `json:"agent_id"`
	TaskID    *string        `json:"task_id"`
	Payload   map[string]any `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > 200 {
				v = 200
			}
			limit = v
		}
	}

	q := `SELECT id, repo_id, type, actor_type, actor_id, agent_id, task_id, payload, created_at
	      FROM events WHERE repo_id = $1`
	args := []any{repoID}
	argIdx := 2

	if before := r.URL.Query().Get("before"); before != "" {
		if t, err := time.Parse(time.RFC3339, before); err == nil {
			q += " AND created_at < $" + itoa(argIdx)
			args = append(args, t)
			argIdx++
		}
	}
	if after := r.URL.Query().Get("after"); after != "" {
		if t, err := time.Parse(time.RFC3339, after); err == nil {
			q += " AND created_at > $" + itoa(argIdx)
			args = append(args, t)
			argIdx++
		}
	}
	if evtType := r.URL.Query().Get("type"); evtType != "" {
		q += " AND type = $" + itoa(argIdx)
		args = append(args, evtType)
		argIdx++
	}
	if agentID := r.URL.Query().Get("agent_id"); agentID != "" {
		q += " AND agent_id = $" + itoa(argIdx)
		args = append(args, agentID)
		argIdx++
	}

	// Fetch one extra to determine has_more.
	q += " ORDER BY created_at DESC LIMIT $" + itoa(argIdx)
	args = append(args, limit+1)
	_ = argIdx

	rows, err := h.db.Query(r.Context(), q, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database error", nil)
		return
	}
	defer rows.Close()

	events := []eventRow{}
	for rows.Next() {
		var e eventRow
		var payload []byte
		if err := rows.Scan(&e.ID, &e.RepoID, &e.Type, &e.ActorType, &e.ActorID,
			&e.AgentID, &e.TaskID, &payload, &e.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scan error", nil)
			return
		}
		json.Unmarshal(payload, &e.Payload)
		if e.Payload == nil {
			e.Payload = map[string]any{}
		}
		events = append(events, e)
	}

	hasMore := false
	if len(events) > limit {
		hasMore = true
		events = events[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"events":   events,
		"has_more": hasMore,
	})
}
