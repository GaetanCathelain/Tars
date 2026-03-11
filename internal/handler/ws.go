package handler

import (
	"log/slog"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, restrict to known origins.
		return true
	},
}

// HandleWS upgrades an HTTP request to a WebSocket connection.
// Authentication is done via a ?token=xxx query parameter because
// browsers cannot set headers on WebSocket handshakes.
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, `{"error":"missing token query parameter"}`, http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateToken(s.JWTSecret, tokenStr)
	if err != nil {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws: upgrade failed", "error", err)
		return
	}

	s.Hub.ServeWS(conn, claims.UserID)
}
