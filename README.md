# Tars

WebUI for orchestrating Claude Code sessions. Create tasks, TARS orchestrates, Claude Code workers execute — with real-time terminal output streaming.

## Quick Start

### Docker Compose (recommended)
```bash
docker compose up --build
```
Open http://localhost:3333

### Manual Development

**Prerequisites:** Go 1.26+, Node.js 22+, PostgreSQL 16

**Backend:**
```bash
# Start PostgreSQL
docker run -d --name tars-db -e POSTGRES_USER=tars -e POSTGRES_PASSWORD=tars_dev -e POSTGRES_DB=tars -p 5432:5432 postgres:16-alpine

# Run backend
export DATABASE_URL="postgres://tars:tars_dev@localhost:5432/tars?sslmode=disable"
go run ./cmd/tars
```

**Frontend (development):**
```bash
cd frontend
npm install
npm run dev
```

## Architecture
- **Backend:** Go + Chi router + pgx (PostgreSQL) + gorilla/websocket
- **Frontend:** SvelteKit 2 + Svelte 5 + Tailwind CSS + xterm.js
- **Auth:** JWT (bcrypt passwords)
- **Real-time:** WebSocket with room-based task subscriptions
- **Workers:** PTY-based Claude Code sessions with live terminal streaming

## API
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/health | Health check |
| POST | /api/auth/register | Register |
| POST | /api/auth/login | Login |
| GET | /api/tasks | List tasks |
| POST | /api/tasks | Create task |
| GET | /api/tasks/:id | Get task |
| GET | /api/tasks/:id/messages | List messages |
| POST | /api/tasks/:id/messages | Send message |
| POST | /api/tasks/:id/workers | Spawn worker |
| GET | /api/workers/:id/output | Replay output |
| DELETE | /api/workers/:id | Kill worker |
| GET | /ws?token=xxx | WebSocket |
