# CHANGES — Backend Foundation (Worker 1)

## What was built

Complete Go backend foundation for the Tars WebUI.

### Structure
- `cmd/tars/main.go` — Entry point with graceful shutdown, config from env
- `embed.go` — Root-level embed for migrations and web assets
- `internal/auth/` — JWT generation/validation (HS256, 24h expiry), bcrypt password hashing, Chi middleware
- `internal/db/` — pgx pool setup, migration runner using golang-migrate with embedded SQL
- `internal/handler/` — HTTP handlers for all API endpoints
- `internal/model/` — Domain types (User, Task, Message, WorkerSession, WorkerOutput)
- `internal/worker/` — Stub package (placeholder for Worker 4)
- `internal/ws/` — Stub package (placeholder for Worker 3)
- `migrations/` — PostgreSQL schema (users, tasks, messages, worker_sessions, worker_output)
- `web/index.html` — Placeholder for embedded frontend

### API Endpoints
| Method | Path | Auth | Status |
|--------|------|------|--------|
| GET | /api/health | No | ✅ |
| POST | /api/auth/register | No | ✅ |
| POST | /api/auth/login | No | ✅ |
| GET | /api/tasks | Yes | ✅ |
| POST | /api/tasks | Yes | ✅ |
| GET | /api/tasks/{id} | Yes | ✅ |
| GET | /api/tasks/{id}/messages | Yes | ✅ |
| POST | /api/tasks/{id}/messages | Yes | ✅ |
| POST | /api/tasks/{id}/workers | Yes | Stub (501) |

### Infrastructure
- `Dockerfile` — Multi-stage build (golang:1.23-alpine → alpine:3.19)
- `docker-compose.yml` — PostgreSQL 16 + app service
- `.env.example` — Config template

### Build status
- `go build ./...` ✅
- `go vet ./...` ✅
