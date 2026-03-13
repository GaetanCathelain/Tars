package api

import (
	"net/http"

	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(queries *db.Queries, hub *ws.Hub, frontendURL string, authHandler *AuthHandler) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware.
	r.Use(Recovery)
	r.Use(RequestID)
	r.Use(Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes (public).
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Get("/github", authHandler.GitHubLogin)
		r.Get("/github/callback", authHandler.GitHubCallback)
		r.Post("/logout", authHandler.Logout)

		// /me requires auth.
		r.Group(func(r chi.Router) {
			r.Use(Auth(queries))
			r.Get("/me", authHandler.Me)
		})
	})

	// Protected API routes.
	r.Group(func(r chi.Router) {
		r.Use(Auth(queries))

		repoHandler := NewRepoHandler(queries)
		r.Route("/api/v1/repos", func(r chi.Router) {
			r.Get("/", repoHandler.List)
			r.Post("/", repoHandler.Create)
			r.Get("/{id}", repoHandler.Get)
			r.Delete("/{id}", repoHandler.Delete)
		})

		taskHandler := NewTaskHandler(queries)
		r.Route("/api/v1/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.List)
			r.Post("/", taskHandler.Create)
			r.Get("/{id}", taskHandler.Get)
			r.Patch("/{id}", taskHandler.Update)
		})
	})

	// WebSocket (auth checked via cookie or query param).
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Check auth via cookie or token query param.
		token := ""
		if cookie, err := r.Cookie("session_token"); err == nil {
			token = cookie.Value
		}
		if t := r.URL.Query().Get("token"); t != "" {
			token = t
		}
		if token == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		session, err := queries.GetSessionByToken(r.Context(), token)
		if err != nil || session == nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		hub.HandleWebSocket(w, r)
	})

	return r
}
