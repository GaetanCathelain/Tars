package handler

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createWorkerRequest struct {
	Prompt string `json:"prompt"`
}

type workerOutputResponse struct {
	ID        int64     `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Data      string    `json:"data"` // base64
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) HandleListWorkers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	rows, err := s.DB.Query(r.Context(),
		`SELECT id, task_id, message_id, status, command, exit_code, started_at, finished_at
		 FROM worker_sessions WHERE task_id = $1 ORDER BY started_at DESC`, taskID)
	if err != nil {
		slog.Error("query workers", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type workerRow struct {
		ID         uuid.UUID  `json:"id"`
		TaskID     uuid.UUID  `json:"task_id"`
		MessageID  *uuid.UUID `json:"message_id,omitempty"`
		Status     string     `json:"status"`
		Command    string     `json:"command"`
		ExitCode   *int       `json:"exit_code,omitempty"`
		StartedAt  time.Time  `json:"started_at"`
		FinishedAt *time.Time `json:"finished_at,omitempty"`
	}

	workers := []workerRow{}
	for rows.Next() {
		var wr workerRow
		if err := rows.Scan(&wr.ID, &wr.TaskID, &wr.MessageID, &wr.Status, &wr.Command, &wr.ExitCode, &wr.StartedAt, &wr.FinishedAt); err != nil {
			slog.Error("scan worker", "error", err)
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		workers = append(workers, wr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

func (s *Server) HandleCreateWorker(w http.ResponseWriter, r *http.Request) {
	if s.WorkerManager == nil {
		writeError(w, "worker manager not configured", http.StatusInternalServerError)
		return
	}

	// Parse task ID
	idStr := chi.URLParam(r, "id")
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	// Validate task exists and belongs to user
	userID := auth.UserIDFromContext(r.Context())
	var ownerID uuid.UUID
	err = s.DB.QueryRow(r.Context(),
		"SELECT created_by FROM tasks WHERE id = $1", taskID,
	).Scan(&ownerID)
	if err != nil {
		writeError(w, "task not found", http.StatusNotFound)
		return
	}
	if ownerID != userID {
		writeError(w, "forbidden", http.StatusForbidden)
		return
	}

	// Parse request
	var req createWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		writeError(w, "prompt required", http.StatusBadRequest)
		return
	}

	// Spawn worker
	session, err := s.WorkerManager.SpawnWorker(r.Context(), taskID, nil, req.Prompt)
	if err != nil {
		slog.Error("spawn worker", "error", err)
		writeError(w, "failed to spawn worker", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": session,
	})
}

func (s *Server) HandleGetWorkerOutput(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid session id", http.StatusBadRequest)
		return
	}

	rows, err := s.DB.Query(r.Context(),
		`SELECT id, session_id, data, created_at FROM worker_output
		 WHERE session_id = $1 ORDER BY id ASC`,
		sessionID,
	)
	if err != nil {
		slog.Error("query worker_output", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	outputs := []workerOutputResponse{}
	for rows.Next() {
		var id int64
		var sid uuid.UUID
		var data []byte
		var createdAt time.Time
		if err := rows.Scan(&id, &sid, &data, &createdAt); err != nil {
			slog.Error("scan worker_output", "error", err)
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		outputs = append(outputs, workerOutputResponse{
			ID:        id,
			SessionID: sid,
			Data:      base64.StdEncoding.EncodeToString(data),
			CreatedAt: createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outputs)
}

func (s *Server) HandleKillWorker(w http.ResponseWriter, r *http.Request) {
	if s.WorkerManager == nil {
		writeError(w, "worker manager not configured", http.StatusInternalServerError)
		return
	}

	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid session id", http.StatusBadRequest)
		return
	}

	if err := s.WorkerManager.KillWorker(sessionID); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "killed"})
}
