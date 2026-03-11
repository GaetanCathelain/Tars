# Worker Engine — Changes

## Added

### PTY-based Worker Engine (`internal/worker/`)

- **`manager.go`** — Worker lifecycle manager with thread-safe session tracking
  - `NewManager(db, hub)` — constructor
  - `SpawnWorker(ctx, taskID, messageID, prompt)` — creates PTY, spawns claude, inserts DB row, broadcasts `worker_start`, starts output capture + process wait goroutines
  - `KillWorker(sessionID)` — terminates a running session
  - `GetSession(sessionID)` / `ActiveSessions()` — query active sessions
  - 15-minute timeout per session (auto-kill)
  - Proper cleanup: close PTY, flush output, update DB, broadcast `worker_end`

- **`pty.go`** — Spawns `claude <prompt>` in a real PTY via `github.com/creack/pty` with xterm-256color, 120x40 terminal size

- **`capture.go`** — Reads PTY output in 4KB chunks, base64-encodes and broadcasts via WebSocket (`worker_output` events), buffers DB writes (flushes every 500ms or 8KB)

### HTTP Handlers (`internal/handler/workers.go`)

- `POST /api/tasks/{id}/workers` — spawn a worker for a task (validates ownership)
- `GET /api/workers/{id}/output` — replay all output chunks for a session (base64-encoded)
- `DELETE /api/workers/{id}` — kill a running worker session

### Integration (`cmd/tars/main.go`)

- Worker manager created and wired into Server struct
- New routes registered under auth middleware

### Dependencies

- Added `github.com/creack/pty v1.1.24`

## Modified

- `internal/handler/auth.go` — Server struct now includes `WorkerManager *worker.Manager`
- `cmd/tars/main.go` — imports worker package, creates manager, adds routes
