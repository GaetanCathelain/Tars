# Changes: Terminal UI Integration

## Summary
Added xterm.js terminal rendering and WebSocket real-time integration to the SvelteKit frontend for live worker output display.

## New Files

### `frontend/src/lib/stores/websocket.svelte.ts`
WebSocket client store managing connection to the Go backend:
- Connects with JWT token via `ws://host/ws?token=xxx`
- Auto-reconnect with exponential backoff (1s → 2s → 4s → max 30s)
- Subscribe/unsubscribe to task rooms
- Dispatches events to messages, workers, and tasks stores
- Connection status tracking: `connecting` | `connected` | `disconnected`

### `frontend/src/lib/stores/workers.svelte.ts`
Worker session store:
- Tracks active/completed workers per task (`Map<taskId, WorkerSession[]>`)
- `spawnWorker(taskId, prompt)` — POST to spawn API
- `killWorker(sessionId)` — DELETE to kill API
- `fetchOutput(sessionId)` — GET for replay data
- Receives `worker_start` and `worker_end` events from WebSocket

### `frontend/src/lib/components/Terminal.svelte`
xterm.js terminal component:
- Read-only terminal with dark theme matching the app (`#0a0a0f` background)
- FitAddon for auto-sizing with ResizeObserver
- WebLinksAddon for clickable URLs
- Exposes `write(data: Uint8Array)` for live output
- Accepts `initialData` for replay of completed sessions
- Custom scrollbar styling

### `frontend/src/lib/components/WorkerCard.svelte`
Collapsible worker session card:
- Header: worker name, status badge (pulsing green dot / ✓ / ✗), duration
- Body: embedded Terminal component with live or replayed output
- Fetches historical output on mount for completed workers
- Subscribes to live WebSocket output for running workers

## Modified Files

### `frontend/src/lib/stores/messages.svelte.ts`
- Added `WorkerEvent` type and `TimelineEntry` union type
- Added `timeline` derived state: messages + worker events sorted chronologically
- Added `addMessage()` for WebSocket-pushed messages
- Added `addWorkerEvent()` for worker start/end markers
- Added mock worker session and events for offline development

### `frontend/src/lib/stores/tasks.svelte.ts`
- Added `updateTaskStatus()` method for WebSocket `task_status` events

### `frontend/src/routes/+layout.svelte`
- Imports and initializes WebSocket store on auth
- Connection status indicator in sidebar footer (green/yellow/red dot)

### `frontend/src/routes/tasks/[id]/+page.svelte`
- Renders timeline (messages + WorkerCards interleaved)
- Subscribes to task room via WebSocket on mount
- Unsubscribes when switching tasks
- WorkerCards appear inline in the conversation flow

## Dependencies Added
- `xterm` — terminal emulator
- `@xterm/addon-fit` — auto-resize terminal to container
- `@xterm/addon-web-links` — clickable links in terminal output

## Build Status
✅ `npm run build` passes cleanly
