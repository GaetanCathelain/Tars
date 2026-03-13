package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GaetanCathelain/Tars/internal/api"
	"github.com/GaetanCathelain/Tars/internal/db"
	"github.com/GaetanCathelain/Tars/internal/ws"
	"github.com/GaetanCathelain/Tars/migrations"
)

func main() {
	// Load config from env.
	databaseURL := envOr("DATABASE_URL", "postgres://tars:tars@localhost:5432/tars?sslmode=disable")
	port := envOr("PORT", "8080")
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectURI := envOr("GITHUB_REDIRECT_URI", "http://localhost:8080/api/v1/auth/github/callback")
	sessionSecret := envOr("SESSION_SECRET", "change-me-in-production")
	frontendURL := envOr("FRONTEND_URL", "http://localhost:5173")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database.
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations.
	if err := db.RunMigrations(ctx, pool, migrations.FS); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Query layer.
	queries := db.NewQueries(pool)

	// WebSocket hub.
	hub := ws.NewHub()
	go hub.Run()

	// Auth handler.
	authHandler := api.NewAuthHandler(queries, githubClientID, githubClientSecret, githubRedirectURI, sessionSecret, frontendURL)

	// Router.
	router := api.NewRouter(queries, hub, frontendURL, authHandler)

	// Server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("TARS backend starting on :%s", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
