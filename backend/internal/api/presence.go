package api

import (
	"net/http"

	"tars/backend/internal/auth"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) getPresence(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	repoID := chi.URLParam(r, "repoId")

	if !h.repoOwnedBy(r, repoID, userID) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "repository not found", nil)
		return
	}

	if h.presence == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"repo_id": repoID,
			"users":   []any{},
		})
		return
	}

	snap := h.presence.Snapshot(repoID)
	writeJSON(w, http.StatusOK, snap)
}
