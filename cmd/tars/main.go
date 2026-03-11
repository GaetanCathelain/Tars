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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	slog.Info("starting tars")

	databaseURL := getEnv("DATABASE_URL", "postgres://tars:tars_dev@localhost:5432/tars?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	port := getEnv("PORT", "8080")

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

	srv := &handler.Server{
		DB:        pool,
		JWTSecret: jwtSecret,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Public routes
	r.Get("/api/health", srv.HandleHealth)
	r.Post("/api/auth/register", srv.HandleRegister)
	r.Post("/api/auth/login", srv.HandleLogin)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Get("/api/tasks", srv.HandleListTasks)
		r.Post("/api/tasks", srv.HandleCreateTask)
		r.Get("/api/tasks/{id}", srv.HandleGetTask)
		r.Get("/api/tasks/{id}/messages", srv.HandleListMessages)
		r.Post("/api/tasks/{id}/messages", srv.HandleCreateMessage)
		r.Post("/api/tasks/{id}/workers", srv.HandleCreateWorker)
	})

	// Static files
	webContent, err := fs.Sub(tars.WebFS, "web")
	if err != nil {
		slog.Error("failed to create sub filesystem", "error", err)
		os.Exit(1)
	}
	r.Handle("/*", http.FileServer(http.FS(webContent)))

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
