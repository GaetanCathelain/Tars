package api

import (
	"net/http"
)

// health returns a simple liveness check.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleWebSocket upgrades the connection to WebSocket after auth validation.
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := h.auth.ValidateSession(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	h.hub.ServeWS(w, r, userID)
}
