# TARS v2

A web-based multiplayer agent conductor — watch, guide, and collaborate with multiple Claude Code agents in real-time. Think Conductor.build, built on open infrastructure.

---

## Overview

TARS v2 lets multiple users open a shared workspace where AI coding agents run inside isolated Git worktrees. Users see live terminal output, assign tasks from a Kanban board, send messages to agents, review diffs, and merge completed work — all simultaneously, with presence awareness showing who is watching what.

### Key Features

- **Multiplayer presence** — see which teammates are viewing which agent
- **Live agent terminal** — xterm.js streaming over WebSocket
- **Task Kanban** — create tasks and assign agents to them
- **Git worktree isolation** — each agent gets its own branch and directory
- **Event timeline** — immutable audit log of everything that happens
- **Diff viewer** — syntax-highlighted git diffs before merging
- **One-click merge** — squash/merge/rebase agent branch into target
- **GitHub OAuth** — auth via GitHub, no passwords

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  Browser (SvelteKit 2 + shadcn-svelte + Tailwind v4)             │
│                                                                  │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │ Auth / OAuth│  │ Kanban Board │  │ Agent Terminal       │   │
│  │ Dashboard   │  │ Task CRUD    │  │ (xterm.js + WS)      │   │
│  └─────────────┘  └──────────────┘  └──────────────────────┘   │
│           │                │                    │               │
│           └────────────────┴────────────────────┘               │
│                        REST + WebSocket                          │
└──────────────────────────────────┬───────────────────────────────┘
                                   │
                    ┌──────────────▼───────────────┐
                    │  Go 1.22+ Backend (chi router) │
                    │                               │
                    │  ┌────────┐  ┌─────────────┐ │
                    │  │ Auth   │  │ WebSocket   │ │
                    │  │(GitHub)│  │ Hub         │ │
                    │  └────────┘  └─────────────┘ │
                    │  ┌────────┐  ┌─────────────┐ │
                    │  │ REST   │  │ Agent       │ │
                    │  │ API    │  │ Process Mgr │ │
                    │  └────────┘  └─────────────┘ │
                    │  ┌────────┐  ┌─────────────┐ │
                    │  │ Git    │  │ Presence    │ │
                    │  │Worktree│  │ Tracker     │ │
                    │  └────────┘  └─────────────┘ │
                    └──────────────┬───────────────┘
                                   │
                    ┌──────────────▼───────────────┐
                    │       PostgreSQL 16           │
                    │  users, repos, tasks, agents, │
                    │  events, sessions, presence   │
                    └──────────────────────────────┘
                                   │
                    ┌──────────────▼───────────────┐
                    │   Claude Code CLI (per agent)  │
                    │   Running in Git worktrees    │
                    │   /workspaces/<repo>/.tars/   │
                    └──────────────────────────────┘
```

---

## Repository Structure

```
tars/
├── backend/               # Go backend
│   ├── cmd/               # main.go entrypoint
│   ├── internal/
│   │   ├── agent/         # Agent process lifecycle
│   │   ├── api/           # HTTP handlers
│   │   ├── auth/          # GitHub OAuth, session
│   │   ├── db/            # pgx database layer
│   │   ├── git/           # Worktree manager
│   │   ├── presence/      # Presence tracker
│   │   ├── task/          # Task service
│   │   └── ws/            # WebSocket hub
│   └── migrations/        # SQL migration files
├── frontend/              # SvelteKit app
│   └── src/
├── shared/                # Contracts (source of truth)
│   ├── api-contracts.md   # REST endpoint specs
│   ├── ws-protocol.md     # WebSocket message specs
│   └── types/             # TypeScript type definitions
│       ├── index.ts
│       ├── models.ts      # Domain models
│       ├── api.ts         # Request/response types
│       └── ws.ts          # WebSocket message types
├── scripts/               # Dev utilities
├── docker-compose.yml
└── Makefile
```

---

## Prerequisites

- **Go** 1.22+
- **Node.js** 20+
- **Docker** + **Docker Compose** v2
- **PostgreSQL** 16 (or use Docker Compose)
- **GitHub OAuth App** — create one at https://github.com/settings/developers
  - Callback URL: `http://localhost:8080/auth/github/callback`

---

## Quick Start

### 1. Clone and configure

```bash
git clone https://github.com/your-org/tars.git
cd tars
cp .env.example .env
```

Edit `.env`:

```env
# Database
DATABASE_URL=postgres://tars:tars@localhost:5432/tars?sslmode=disable

# GitHub OAuth
GITHUB_CLIENT_ID=your_client_id
GITHUB_CLIENT_SECRET=your_client_secret

# Session
SESSION_SECRET=change-me-in-production-32-chars-min

# Server
PORT=8080
FRONTEND_URL=http://localhost:5173
```

### 2. Start dependencies

```bash
docker compose up -d postgres
```

### 3. Run the backend

```bash
cd backend
go run ./cmd/tars
# Server starts on :8080
```

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev
# App starts on :5173
```

### 5. Open

Navigate to `http://localhost:5173` and sign in with GitHub.

---

## Development with Docker Compose (full stack)

```bash
make up        # Start everything (postgres + backend + frontend)
make down      # Stop everything
make logs      # Tail all logs
make migrate   # Run database migrations
```

---

## API Reference

Full REST API specification: [`shared/api-contracts.md`](shared/api-contracts.md)

Full WebSocket protocol specification: [`shared/ws-protocol.md`](shared/ws-protocol.md)

Shared TypeScript types: [`shared/types/`](shared/types/)

### Quick Reference

| Resource    | Base path                        |
|-------------|----------------------------------|
| Auth        | `/auth/github/...`               |
| Repos       | `/api/v1/repos`                  |
| Tasks       | `/api/v1/repos/:id/tasks`        |
| Agents      | `/api/v1/repos/:id/agents`       |
| Presence    | `/api/v1/repos/:id/presence`     |
| Events      | `/api/v1/repos/:id/events`       |
| Diffs       | `/api/v1/repos/:id/agents/:id/diff` |
| WebSocket   | `ws://localhost:8080/ws`         |

---

## Tech Stack

| Layer      | Technology                                      |
|------------|-------------------------------------------------|
| Frontend   | SvelteKit 2, shadcn-svelte, Tailwind v4, xterm.js |
| Backend    | Go 1.22+, chi router, gorilla/websocket, pgx    |
| Database   | PostgreSQL 16                                   |
| Auth       | GitHub OAuth 2.0, HttpOnly session cookies      |
| Agents     | Claude Code CLI                                 |
| Isolation  | Git worktrees (one per agent)                   |
| Infra      | Docker Compose                                  |

---

## Contributing

1. Branch from `main`.
2. Follow the contracts in `shared/` — do not change API shapes without updating contracts first.
3. Run `make test` before opening a PR.
