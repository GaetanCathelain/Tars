package api

import (
	"net/http"

	"tars/backend/internal/agent"
	"tars/backend/internal/auth"
	"tars/backend/internal/db"
	"tars/backend/internal/git"
	"tars/backend/internal/presence"
	"tars/backend/internal/ws"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Handler holds all dependencies for HTTP handlers.
type Handler struct {
	auth     *auth.Manager
	github   *auth.GitHub
	hub      *ws.Hub
	db       *db.Pool
	agents   *agent.Manager
	worktree *git.WorktreeManager
	presence *presence.Tracker
}

// Config holds router configuration.
type Config struct {
	Auth           *auth.Manager
	GitHub         *auth.GitHub
	Hub            *ws.Hub
	DB             *db.Pool
	Agents         *agent.Manager
	Worktree       *git.WorktreeManager
	Presence       *presence.Tracker
	AllowedOrigins string
}

// NewRouter builds and returns the chi router with all routes registered.
func NewRouter(cfg Config) http.Handler {
	h := &Handler{
		auth:     cfg.Auth,
		github:   cfg.GitHub,
		hub:      cfg.Hub,
		db:       cfg.DB,
		agents:   cfg.Agents,
		worktree: cfg.Worktree,
		presence: cfg.Presence,
	}

	r := chi.NewRouter()

	// Global middleware stack.
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(CORSMiddleware(cfg.AllowedOrigins))

	// Health check — no auth.
	r.Get("/health", h.health)

	// WebSocket — auth handled in handler (must validate before upgrade).
	r.Get("/ws", h.handleWebSocket)

	// API v1 routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes — no session required.
		r.Route("/auth", func(r chi.Router) {
			r.Get("/github/login", h.handleGitHubLogin)
			r.Get("/github/callback", h.handleGitHubCallback)
			r.Post("/register", h.handleRegister)
			r.Post("/login", h.handleLogin)

			// Protected auth routes.
			r.Group(func(r chi.Router) {
				r.Use(cfg.Auth.RequireAuth)
				r.Get("/me", h.handleMe)
				r.Post("/logout", h.handleLogout)
			})
		})

		// All remaining routes require authentication.
		r.Group(func(r chi.Router) {
			r.Use(cfg.Auth.RequireAuth)

			// Repositories.
			r.Route("/repos", func(r chi.Router) {
				r.Get("/", h.listRepos)
				r.Post("/", h.createRepo)

				r.Route("/{repoId}", func(r chi.Router) {
					r.Get("/", h.getRepo)
					r.Patch("/", h.updateRepo)
					r.Delete("/", h.deleteRepo)

					// Tasks (scoped to repo).
					r.Route("/tasks", func(r chi.Router) {
						r.Get("/", h.listTasks)
						r.Post("/", h.createTask)
						r.Route("/{taskId}", func(r chi.Router) {
							r.Get("/", h.getTask)
							r.Patch("/", h.updateTask)
							r.Delete("/", h.deleteTask)
						})
					})

					// Agents (scoped to repo).
					r.Route("/agents", func(r chi.Router) {
						r.Get("/", h.listAgents)
						r.Post("/", h.spawnAgent)
						r.Route("/{agentId}", func(r chi.Router) {
							r.Get("/", h.getAgent)
							r.Post("/stop", h.stopAgent)
							r.Post("/input", h.sendAgentInput)
							r.Get("/logs", h.getAgentLogs)
							r.Post("/merge", h.mergeAgent)
							r.Get("/diff", h.getAgentDiff)
						})
					})

					// Presence.
					r.Get("/presence", h.getPresence)

					// Events / timeline.
					r.Get("/events", h.listEvents)
				})
			})
		})
	})

	return r
}
