package api

import (
	"encoding/json"
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RepoHandler struct {
	queries *db.Queries
}

func NewRepoHandler(queries *db.Queries) *RepoHandler {
	return &RepoHandler{queries: queries}
}

func (h *RepoHandler) List(w http.ResponseWriter, r *http.Request) {
	repos, err := h.queries.ListRepos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list repos")
		return
	}
	if repos == nil {
		repos = []models.Repo{}
	}
	writeJSON(w, http.StatusOK, repos)
}

func (h *RepoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string  `json:"name"`
		URL           string  `json:"url"`
		LocalPath     *string `json:"local_path"`
		DefaultBranch *string `json:"default_branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name == "" || input.URL == "" {
		writeError(w, http.StatusBadRequest, "name and url are required")
		return
	}

	user := UserFromContext(r.Context())
	repo := &models.Repo{
		Name:          input.Name,
		URL:           input.URL,
		LocalPath:     input.LocalPath,
		DefaultBranch: "main",
		AddedBy:       &user.ID,
	}
	if input.DefaultBranch != nil {
		repo.DefaultBranch = *input.DefaultBranch
	}

	if err := h.queries.CreateRepo(r.Context(), repo); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create repo")
		return
	}
	writeJSON(w, http.StatusCreated, repo)
}

func (h *RepoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repo id")
		return
	}

	repo, err := h.queries.GetRepoByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get repo")
		return
	}
	if repo == nil {
		writeError(w, http.StatusNotFound, "repo not found")
		return
	}
	writeJSON(w, http.StatusOK, repo)
}

func (h *RepoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid repo id")
		return
	}

	if err := h.queries.DeleteRepo(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "repo not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
