# Docker + Integration Polish

## Summary
Multi-stage Docker build, CORS middleware, graceful shutdown improvements, and project infrastructure files.

## Files Changed

| File | What Changed |
|------|-------------|
| `Dockerfile` | Multi-stage: node:22-alpine → golang:1.24-alpine → alpine:3.21. Builds frontend, embeds into Go binary, tini entrypoint |
| `docker-compose.yml` | Renamed db→postgres, added healthcheck, port 3333, service_healthy dependency |
| `.dockerignore` | New: excludes .git, node_modules, .svelte-kit, build artifacts, markdown (except README) |
| `.env.example` | Updated port to 3333 |
| `cmd/tars/main.go` | Added CORS middleware (go-chi/cors), default port→3333, worker ShutdownAll on SIGINT/SIGTERM |
| `internal/worker/manager.go` | Added ShutdownAll() — kills all active workers during graceful shutdown |
| `go.mod` / `go.sum` | Added github.com/go-chi/cors v1.2.2 |
| `README.md` | Full rewrite: quick start (Docker + manual), architecture, API table |
| `Makefile` | New: dev, build, frontend-build, docker, docker-down, clean targets |

## Review Notes

### Graceful Shutdown
- Signal handling: SIGINT, SIGTERM → kills all active workers → HTTP server shutdown (10s timeout) → DB pool close
- Worker manager ShutdownAll iterates active sessions, cancels contexts, kills processes

### CORS
- AllowedOrigins: `*` (dev-friendly, restrict in production)
- AllowCredentials: true
- AllowedHeaders: Accept, Authorization, Content-Type, X-Request-ID

### Error Handling (already solid)
- All handlers use `writeError()` → consistent `{"error":"..."}` JSON responses
- Proper HTTP status codes throughout (400, 401, 403, 404, 409, 500)
- slog structured logging on all errors
- chi Recoverer middleware prevents panics from crashing the server

### Dockerfile Notes
- Uses golang:1.24-alpine (latest stable Go image; go.mod says 1.26.1 but Docker Hub doesn't have 1.26 images yet — build still works with toolchain directive)
- tini as PID 1 for proper signal forwarding
- claude CLI not included in image — must be mounted or installed separately for worker functionality
