# Backend Changes — TARS v2 Phase 1

## What Was Built

Complete Go backend for TARS v2 Phase 1 foundation layer.

### Structure
```
backend/
├── cmd/server/main.go           # Entry point, graceful shutdown
├── internal/
│   ├── api/
│   │   ├── router.go            # Chi router, all routes wired
│   │   ├── auth.go              # GitHub OAuth (login, callback, logout, me)
│   │   ├── repos.go             # Repo CRUD (list, create, get, delete)
│   │   ├── tasks.go             # Task CRUD (list, create, get, update)
│   │   ├── middleware.go        # RequestID, Logger, Recovery, Auth
│   │   └── helpers.go           # JSON response helpers
│   ├── ws/
│   │   └── hub.go               # WebSocket hub, client management, broadcast
│   ├── db/
│   │   ├── db.go                # pgx connection pool
│   │   ├── migrations.go        # Migration runner (embed.FS based)
│   │   └── queries.go           # All SQL queries (parameterized)
│   └── models/
│       └── models.go            # Go structs for all 7 tables
├── migrations/
│   ├── embed.go                 # Embeds SQL files
│   ├── 001_initial.up.sql       # 7 tables + 10 indexes
│   └── 001_initial.down.sql     # Clean teardown
├── go.mod
└── go.sum
```

### API Endpoints
- `GET /health` — health check
- `GET /api/v1/auth/github` — GitHub OAuth redirect
- `GET /api/v1/auth/github/callback` — OAuth callback, upsert user, create session
- `POST /api/v1/auth/logout` — invalidate session
- `GET /api/v1/auth/me` — current user (authed)
- `GET /api/v1/repos` — list repos (authed)
- `POST /api/v1/repos` — create repo (authed)
- `GET /api/v1/repos/{id}` — get repo (authed)
- `DELETE /api/v1/repos/{id}` — delete repo (authed)
- `GET /api/v1/tasks` — list tasks with filters (authed)
- `POST /api/v1/tasks` — create task (authed)
- `GET /api/v1/tasks/{id}` — get task (authed)
- `PATCH /api/v1/tasks/{id}` — update task (authed)
- `GET /ws` — WebSocket upgrade (authed via cookie or token param)

### Infrastructure
- `docker-compose.yml` — PostgreSQL 16 with health check
- `Dockerfile.backend` — Multi-stage build (golang:1.22-alpine → alpine:3.19)

### Key Decisions
- Migrations embedded via `migrations/embed.go` package (avoids `..` path issues with `go:embed`)
- Session tokens generated with HMAC-SHA256 using `SESSION_SECRET`
- Dynamic query building for task updates (only set provided fields)
- Consistent JSON envelope: `{"data":...}` / `{"error":"..."}`
- CORS configured for `FRONTEND_URL` with credentials

### Status
- ✅ Compiles clean (`go build ./...`)
- ✅ Passes vet (`go vet ./...`)
- ✅ 7 commits on `tars/v2-phase1-foundation`
