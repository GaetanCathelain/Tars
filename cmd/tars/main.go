package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tars "github.com/GaetanCathelain/Tars"
	"github.com/GaetanCathelain/Tars/internal/auth"
	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/handler"
	"github.com/GaetanCathelain/Tars/internal/model"
	"github.com/GaetanCathelain/Tars/internal/worker"
	"github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	slog.Info("starting tars")

	databaseURL := getEnv("DATABASE_URL", "postgres://tars:tars_dev@localhost:5432/tars?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	port := getEnv("PORT", "3333")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run migrations
	migFS, err := fs.Sub(tars.MigrationsFS, "migrations")
	if err != nil {
		slog.Error("failed to create migrations sub fs", "error", err)
		os.Exit(1)
	}
	if err := db.RunMigrations(databaseURL, migFS); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Connect to database
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	hub := ws.NewHub()
	hub.OnMessage = makeMessageHandler(pool)
	go hub.Run()

	mgr := worker.NewManager(pool, hub)

	srv := &handler.Server{
		DB:            pool,
		JWTSecret:     jwtSecret,
		Hub:           hub,
		WorkerManager: mgr,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// CORS — permissive for development
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Get("/api/health", srv.HandleHealth)
	r.Post("/api/auth/register", srv.HandleRegister)
	r.Post("/api/auth/login", srv.HandleLogin)

	// WebSocket (auth via query param, not middleware)
	r.Get("/ws", srv.HandleWS)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Get("/api/tasks", srv.HandleListTasks)
		r.Post("/api/tasks", srv.HandleCreateTask)
		r.Get("/api/tasks/{id}", srv.HandleGetTask)
		r.Get("/api/tasks/{id}/messages", srv.HandleListMessages)
		r.Post("/api/tasks/{id}/messages", srv.HandleCreateMessage)
		r.Post("/api/tasks/{id}/workers", srv.HandleCreateWorker)
		r.Get("/api/workers/{id}/output", srv.HandleGetWorkerOutput)
		r.Delete("/api/workers/{id}", srv.HandleKillWorker)
	})

	// SPA static files with fallback to index.html
	webContent, err := fs.Sub(tars.WebFS, "web")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		os.Exit(1)
	}
	staticFS := http.FS(webContent)
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		f, err := webContent.Open(path[1:]) // strip leading /
		if err == nil {
			f.Close()
			http.FileServer(staticFS).ServeHTTP(w, r)
			return
		}
		// Fallback: serve index.html for SPA routing
		r.URL.Path = "/"
		http.FileServer(staticFS).ServeHTTP(w, r)
	}))

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig)

		// Kill active workers first
		mgr.ShutdownAll()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
		cancel()
	}()

	slog.Info("listening", "port", port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// makeMessageHandler returns a ws.MessageHandler that persists messages to the database.
func makeMessageHandler(pool *pgxpool.Pool) ws.MessageHandler {
	return func(userID uuid.UUID, taskID uuid.UUID, content string) (interface{}, error) {
		msg := model.Message{
			TaskID:     taskID,
			SenderType: "user",
			SenderID:   &userID,
			Content:    content,
		}

		err := pool.QueryRow(context.Background(),
			"INSERT INTO messages (task_id, sender_type, sender_id, content) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
			msg.TaskID, msg.SenderType, msg.SenderID, msg.Content,
		).Scan(&msg.ID, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}

		return msg, nil
	}
}
