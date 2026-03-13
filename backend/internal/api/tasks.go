package api

import (
	"encoding/json"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskHandler struct {
	queries *db.Queries
}

func NewTaskHandler(queries *db.Queries) *TaskHandler {
	return &TaskHandler{queries: queries}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	var repoID *uuid.UUID
	if rid := r.URL.Query().Get("repo_id"); rid != "" {
		parsed, err := uuid.Parse(rid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid repo_id")
			return
		}
		repoID = &parsed
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	tasks, err := h.queries.ListTasks(r.Context(), repoID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RepoID      *uuid.UUID `json:"repo_id"`
		Title       string     `json:"title"`
		Description *string    `json:"description"`
		Priority    *int       `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	user := UserFromContext(r.Context())
	task := &models.Task{
		RepoID:      input.RepoID,
		Title:       input.Title,
		Description: input.Description,
		Status:      "pending",
		CreatedBy:   &user.ID,
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}

	if err := h.queries.CreateTask(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.queries.GetTaskByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get task")
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var input struct {
		Status      *string `json:"status"`
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Priority    *int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate status if provided.
	if input.Status != nil {
		valid := map[string]bool{
			"pending": true, "assigned": true, "in_progress": true,
			"review": true, "done": true, "failed": true,
		}
		if !valid[*input.Status] {
			writeError(w, http.StatusBadRequest, "invalid status value")
			return
		}
	}

	task, err := h.queries.UpdateTask(r.Context(), id, input.Status, input.Title, input.Description, input.Priority)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, task)
}
