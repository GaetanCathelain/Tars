package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/GaetanCathelain/Tars/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createTaskRequest struct {
	Title string `json:"title"`
}

func (s *Server) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(r.Context(),
		"SELECT id, title, status, created_by, created_at, updated_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		slog.Error("list tasks", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			slog.Error("scan task", "error", err)
			writeError(w, "internal error", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		writeError(w, "title required", http.StatusBadRequest)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	task := model.Task{
		Title:     req.Title,
		Status:    "open",
		CreatedBy: userID,
	}

	err := s.DB.QueryRow(r.Context(),
		"INSERT INTO tasks (title, status, created_by) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at",
		task.Title, task.Status, task.CreatedBy,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		slog.Error("create task", "error", err)
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid task id", http.StatusBadRequest)
		return
	}

	var t model.Task
	err = s.DB.QueryRow(r.Context(),
		"SELECT id, title, status, created_by, created_at, updated_at FROM tasks WHERE id = $1",
		id,
	).Scan(&t.ID, &t.Title, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		writeError(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// mustParseUUID parses a UUID string, returning uuid.Nil on failure.
func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
