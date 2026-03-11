# Changes: Live Terminal Output in Worker Cards

## Summary
Wired up real-time terminal output rendering in the TARS WebUI using xterm.js. Worker cards now show live ANSI terminal output streamed via WebSocket, and historical output for completed workers loaded via REST API.

## New Files
- **`frontend/src/lib/components/app/terminal-output.svelte`** — xterm.js Terminal wrapper component
  - Read-only terminal with dark theme matching the app
  - Fetches historical output via `GET /api/workers/{id}/output` for completed workers
  - Subscribes to live WebSocket output for running workers
  - Auto-fits to container with ResizeObserver
  - Min 200px / max 400px height

## Modified Files

### `frontend/src/lib/stores/websocket.svelte.ts`
- Added `onmessage` handler that parses incoming JSON messages
- Handles `worker_output` (base64 decode → dispatch to subscribers), `worker_start`, `worker_end`, `task_status`, and `message` events
- Added `subscribe(sessionId, callback)` / `unsubscribe(sessionId, callback)` for worker output streaming
- Added auto-reconnect on disconnect (3s delay)
- Re-subscribes to task on reconnect

### `frontend/src/lib/stores/workers.svelte.ts`
- Added `addWorker(worker)` — inserts new worker from WebSocket `worker_start` events
- Added `updateWorkerStatus(sessionId, status, exitCode?)` — updates worker status from `worker_end` events
- Removed mock mode / mock data

### `frontend/src/lib/stores/tasks.svelte.ts`
- Added `updateTaskStatus(taskId, status)` — updates task status from WebSocket `task_status` events
- Removed mock mode / mock data

### `frontend/src/lib/stores/messages.svelte.ts`
- Added `addMessage(message)` — appends incoming WebSocket messages for the current task
- Deduplicates by message ID
- Removed mock mode / mock data

### `frontend/src/lib/components/app/worker-card.svelte`
- Replaced "Terminal output placeholder" with `terminal-output` component
- Made card collapsible (click header to toggle terminal visibility)
- Shows actual worker status via Badge component
- Displays truncated worker session ID

### `frontend/src/lib/components/app/chat-view.svelte`
- Added WebSocket connection on auth (auto-connects when token available)
- Subscribes to task's WebSocket channel when task is selected
- Unsubscribes when switching tasks (cleanup via `$effect` return)
- Added WebSocket connection indicator (green/gray dot) in header

## Dependencies Added
- `@xterm/xterm` — terminal emulator
- `@xterm/addon-fit` — auto-fit terminal to container
- `@xterm/addon-web-links` — clickable URLs in terminal output
