package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/GaetanCathelain/Tars/internal/model"
	"github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createMessageRequest struct {
	Content string `json:"content"`
}

func (s *Server) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		writeError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	rows, err := s.DB.Query(r.Context(),
		"SELECT id, task_id, sender_type, sender_id, content, created_at FROM messages WHERE task_id = $1 ORDER BY created_at ASC",
		taskID)
	if err != nil {
		slog.Error("list messages", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []model.Message{}
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.TaskID, &m.SenderType, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			slog.Error("scan message", "error", err)
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		messages = append(messages, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (s *Server) HandleCreateMessage(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		writeError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		writeError(w, "content required", http.StatusBadRequest)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	msg := model.Message{
		TaskID:     taskID,
		SenderType: "user",
		SenderID:   &userID,
		Content:    req.Content,
	}

	err = s.DB.QueryRow(r.Context(),
		"INSERT INTO messages (task_id, sender_type, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
		msg.TaskID, msg.SenderType, msg.SenderID, msg.Content,
	).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		slog.Error("create message", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Broadcast to WebSocket subscribers
	if s.Hub != nil {
		s.Hub.BroadcastToTask(msg.TaskID, &ws.OutgoingMessage{
			Type:    "message",
			TaskID:  msg.TaskID,
			Message: msg,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}
