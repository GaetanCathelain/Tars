package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"tars/backend/internal/agent"
	"tars/backend/internal/api"
	"tars/backend/internal/auth"
	"tars/backend/internal/db"
	"tars/backend/internal/git"
	"tars/backend/internal/presence"
	"tars/backend/internal/ws"
)

func main() {
	// Support --migrate-only flag for running migrations without starting the server.
	if len(os.Args) > 1 && os.Args[1] == "--migrate-only" {
		if err := runMigrationsOnly(); err != nil {
			log.Fatalf("migrations: %v", err)
		}
		return
	}

	if err := run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func runMigrationsOnly() error {
	databaseURL := requireEnv("DATABASE_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	migrationsDir := migrationsPath()
	if err := db.RunMigrations(context.Background(), pool, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Println("migrations: complete")
	return nil
}

func run() error {
	cfg := loadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect to PostgreSQL.
	pool, err := db.Connect(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	log.Println("database: connected")

	// Run migrations.
	migrationsDir := migrationsPath()
	if err := db.RunMigrations(context.Background(), pool, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Println("database: migrations up to date")

	// Build auth manager.
	authMgr := auth.New(auth.Config{
		SessionSecret: []byte(cfg.sessionSecret),
		CookieSecure:  cfg.cookieSecure,
	})

	// Build GitHub OAuth handler.
	ghOAuth := auth.NewGitHub(auth.GitHubConfig{
		ClientID:     cfg.githubClientID,
		ClientSecret: cfg.githubClientSecret,
	}, authMgr)

	// Build WebSocket hub.
	hub := ws.New()
	go hub.Run()

	// Build agent infrastructure.
	agentMgr := agent.NewManager(pool, hub)
	worktreeMgr := git.NewWorktreeManager()
	presenceTracker := presence.New(hub)

	// Build router.
	router := api.NewRouter(api.Config{
		Auth:           authMgr,
		GitHub:         ghOAuth,
		Hub:            hub,
		DB:             pool,
		Agents:         agentMgr,
		Worktree:       worktreeMgr,
		Presence:       presenceTracker,
		AllowedOrigins: cfg.allowedOrigins,
	})

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server: listening on :%s", cfg.port)
		serverErr <- srv.ListenAndServe()
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Printf("server: received signal %s, shutting down", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Println("server: shutdown complete")
	return nil
}

type config struct {
	port             string
	databaseURL      string
	githubClientID   string
	githubClientSecret string
	sessionSecret    string
	allowedOrigins   string
	cookieSecure     bool
}

func loadConfig() config {
	cfg := config{
		port:             getEnv("PORT", "8080"),
		databaseURL:      requireEnv("DATABASE_URL"),
		githubClientID:   getEnv("GITHUB_CLIENT_ID", ""),
		githubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		sessionSecret:    requireEnv("JWT_SECRET"),
		allowedOrigins:   getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
		cookieSecure:     parseBool(getEnv("COOKIE_SECURE", "false")),
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// migrationsPath returns the absolute path to the migrations directory,
// resolved relative to the binary's source location for dev and Docker.
func migrationsPath() string {
	// In production/Docker the binary sits at /app/server and migrations at /app/migrations.
	if p := os.Getenv("MIGRATIONS_DIR"); p != "" {
		return p
	}
	// During development: resolve relative to this source file.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "migrations"
	}
	// filename = .../backend/cmd/server/main.go → go up 3 levels → backend/
	return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
}
