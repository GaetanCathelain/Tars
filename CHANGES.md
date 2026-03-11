# Changes — WebSocket + Real-time Layer

## New Files
- **`internal/ws/hub.go`** — WebSocket hub with room-based broadcasting
- **`internal/handler/ws.go`** — HTTP→WS upgrade handler with JWT query param auth

## Modified Files
- **`internal/handler/auth.go`** — Added `Hub *ws.Hub` field to `Server` struct
- **`internal/handler/messages.go`** — `HandleCreateMessage` now broadcasts to WS subscribers
- **`cmd/tars/main.go`** — Creates hub, wires message handler, adds `/ws` route
- **`go.mod` / `go.sum`** — Added `github.com/gorilla/websocket v1.5.3`

## Architecture

### Hub (`internal/ws/hub.go`)
- **Hub** — central goroutine processing register/unregister/broadcast channels
- **Client** — wraps a `gorilla/websocket.Conn` with read/write pumps
- **Rooms** — `map[uuid.UUID]map[*Client]bool` for task-scoped broadcast
- **`BroadcastToTask(taskID, msg)`** — public API for broadcasting from any handler
- **`OnMessage` callback** — pluggable handler for persisting chat messages from WS clients
- Ping/pong: 30s ping interval, 60s pong timeout
- Thread-safe: `sync.RWMutex` on rooms/clients maps

### Protocol
Client → Server: `subscribe`, `unsubscribe`, `message`
Server → Client: `message`, `worker_start`, `worker_output`, `worker_end`, `task_status`

### Auth
WebSocket connections authenticate via `?token=<jwt>` query parameter (browsers can't set headers on WS handshake). Token is validated using the existing `auth.ValidateToken()`.

### Integration
- REST `POST /api/tasks/{id}/messages` broadcasts via hub (both REST and WS clients see new messages)
- WS `{"type":"message"}` persists to DB then broadcasts (same path, different entry point)
