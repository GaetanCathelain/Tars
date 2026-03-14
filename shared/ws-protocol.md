# TARS v2 — WebSocket Protocol

> **Protocol version**: 1.0.0
> **Endpoint**: `ws://localhost:8080/ws`
> **Auth**: The same `tars_session` cookie used by the REST API. The server rejects the upgrade with `401` if the cookie is absent or invalid.

---

## Table of Contents

1. [Connection Lifecycle](#1-connection-lifecycle)
2. [Envelope Format](#2-envelope-format)
3. [Client → Server Messages](#3-client--server-messages)
4. [Server → Client Messages](#4-server--client-messages)
5. [Channel Subscriptions](#5-channel-subscriptions)
6. [Error Handling](#6-error-handling)
7. [Heartbeat / Keepalive](#7-heartbeat--keepalive)
8. [Message Reference Table](#8-message-reference-table)

---

## 1. Connection Lifecycle

```
Client                              Server
  |                                   |
  |-- HTTP Upgrade (cookie auth) ---> |
  |<-- 101 Switching Protocols ------|
  |                                   |
  |-- subscribe (repo channel) -----> |
  |<-- subscribed ------------------- |
  |                                   |
  |<-- presence.update -------------- |  (server pushes current presence)
  |                                   |
  |   ... normal message exchange ... |
  |                                   |
  |-- ping -------------------------> |
  |<-- pong ------------------------- |
  |                                   |
  |-- unsubscribe ------------------> |
  |   (or TCP close)                  |
  |<-- [server removes presence] ---- |
```

**One WebSocket connection per browser tab.** A single connection can subscribe to multiple channels simultaneously.

---

## 2. Envelope Format

Every message (both directions) is a JSON object with a `type` discriminator and a `payload` object.

```ts
interface Envelope {
  type: string;          // message type identifier (see below)
  id?: string;           // optional client-generated request ID for correlation
  channel?: string;      // channel this message belongs to (if applicable)
  payload: object;       // message-specific data
}
```

- `type` — always present, identifies the message kind.
- `id` — optional, set by client for request/response correlation; server echoes it in the response.
- `channel` — present when a message is scoped to a subscription channel.
- `payload` — message-specific structured data (never `null`; use `{}` if empty).

**Example envelope**:

```json
{
  "type": "subscribe",
  "id": "req_01",
  "payload": {
    "channel": "repo:repo_01J..."
  }
}
```

---

## 3. Client → Server Messages

### 3.1 `subscribe`

Subscribe to a channel to receive its server-push events.

```json
{
  "type": "subscribe",
  "id": "req_01",
  "payload": {
    "channel": "repo:repo_01J..."
  }
}
```

**Payload fields**:

| Field     | Type   | Required | Description          |
|-----------|--------|----------|----------------------|
| `channel` | string | yes      | Channel identifier   |

**Server responds with**: `subscribed` or `error`.

---

### 3.2 `unsubscribe`

Unsubscribe from a channel.

```json
{
  "type": "unsubscribe",
  "id": "req_02",
  "payload": {
    "channel": "repo:repo_01J..."
  }
}
```

**Server responds with**: `unsubscribed` or `error`.

---

### 3.3 `presence.update`

Announce what the user is currently viewing. Sent by the client whenever navigation changes.

```json
{
  "type": "presence.update",
  "payload": {
    "repo_id": "repo_01J...",
    "viewing_agent_id": "agent_01J..."
  }
}
```

**Payload fields**:

| Field              | Type         | Required | Description                                           |
|--------------------|--------------|----------|-------------------------------------------------------|
| `repo_id`          | string       | yes      | Currently viewed repo                                 |
| `viewing_agent_id` | string\|null | no       | Agent terminal the user has open, or `null` if none  |

**Side effect**: Server broadcasts `presence.snapshot` to all subscribers of `repo:{repoId}`.

---

### 3.4 `agent.input`

Send text input to a running agent's stdin. Alternative to the REST `POST /agents/:id/input` for lower-latency interactive use.

```json
{
  "type": "agent.input",
  "payload": {
    "agent_id": "agent_01J...",
    "text": "Focus on the auth module first."
  }
}
```

**Payload fields**:

| Field      | Type   | Required | Description                       |
|------------|--------|----------|-----------------------------------|
| `agent_id` | string | yes      | Target agent ID                   |
| `text`     | string | yes      | Text to write to the agent's stdin |

**Server responds with**: `error` if agent is not running; no response on success (output will stream back via `agent.output`).

---

### 3.5 `ping`

Heartbeat ping from client.

```json
{
  "type": "ping",
  "payload": {}
}
```

**Server responds with**: `pong`.

---

## 4. Server → Client Messages

### 4.1 `subscribed`

Confirmation that a channel subscription was accepted.

```json
{
  "type": "subscribed",
  "id": "req_01",
  "payload": {
    "channel": "repo:repo_01J..."
  }
}
```

---

### 4.2 `unsubscribed`

Confirmation that a channel subscription was removed.

```json
{
  "type": "unsubscribed",
  "id": "req_02",
  "payload": {
    "channel": "repo:repo_01J..."
  }
}
```

---

### 4.3 `agent.output`

A chunk of stdout/stderr from a running agent. High-frequency — may arrive many times per second.

```json
{
  "type": "agent.output",
  "channel": "agent:agent_01J...",
  "payload": {
    "agent_id": "agent_01J...",
    "seq": 42,
    "ts": "2024-01-15T10:30:01.123Z",
    "stream": "stdout",
    "text": "Analyzing the codebase..."
  }
}
```

**Payload fields**:

| Field      | Type    | Description                                      |
|------------|---------|--------------------------------------------------|
| `agent_id` | string  | Agent that produced this output                  |
| `seq`      | integer | Monotonically increasing sequence number (per agent) for ordering |
| `ts`       | string  | ISO 8601 timestamp with millisecond precision    |
| `stream`   | string  | `stdout` or `stderr`                             |
| `text`     | string  | Raw text chunk (may contain ANSI escape codes)   |

---

### 4.4 `agent.status`

Broadcast when an agent's status changes (started, stopped, crashed).

```json
{
  "type": "agent.status",
  "channel": "repo:repo_01J...",
  "payload": {
    "agent_id": "agent_01J...",
    "status": "stopped",
    "exit_code": 0,
    "ts": "2024-01-15T12:00:00Z"
  }
}
```

**Payload fields**:

| Field       | Type         | Description                                            |
|-------------|--------------|--------------------------------------------------------|
| `agent_id`  | string       | Agent that changed status                              |
| `status`    | string       | New status: `starting`, `running`, `stopped`, `crashed`|
| `exit_code` | integer\|null| Process exit code (present when `stopped` or `crashed`)|
| `ts`        | string       | ISO 8601 timestamp                                     |

---

### 4.5 `task.updated`

Broadcast when any task field changes (including status or agent assignment).

```json
{
  "type": "task.updated",
  "channel": "repo:repo_01J...",
  "payload": {
    "task": {
      "id": "task_01J...",
      "repo_id": "repo_01J...",
      "title": "Implement OAuth login",
      "description": "...",
      "status": "done",
      "priority": 2,
      "agent_id": "agent_01J...",
      "created_by": "usr_01J...",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T12:00:00Z"
    }
  }
}
```

---

### 4.6 `task.created`

Broadcast when a new task is created.

```json
{
  "type": "task.created",
  "channel": "repo:repo_01J...",
  "payload": {
    "task": { "...full task object..." }
  }
}
```

---

### 4.7 `task.deleted`

Broadcast when a task is deleted.

```json
{
  "type": "task.deleted",
  "channel": "repo:repo_01J...",
  "payload": {
    "task_id": "task_01J..."
  }
}
```

---

### 4.8 `presence.snapshot`

Full current presence state for a repo. Sent to a client immediately after subscribing to a `repo:` channel, and re-broadcast to all subscribers whenever any user's presence changes.

```json
{
  "type": "presence.snapshot",
  "channel": "repo:repo_01J...",
  "payload": {
    "repo_id": "repo_01J...",
    "users": [
      {
        "user_id": "usr_01J...",
        "login": "octocat",
        "avatar_url": "https://avatars.githubusercontent.com/u/583231",
        "viewing_agent_id": "agent_01J...",
        "last_seen": "2024-01-15T12:00:00Z"
      }
    ]
  }
}
```

---

### 4.9 `event.created`

Broadcast when a new timeline event is recorded.

```json
{
  "type": "event.created",
  "channel": "repo:repo_01J...",
  "payload": {
    "event": {
      "id": "evt_01J...",
      "repo_id": "repo_01J...",
      "type": "agent.spawned",
      "actor_type": "user",
      "actor_id": "usr_01J...",
      "agent_id": "agent_01J...",
      "task_id": null,
      "payload": {
        "name": "claude-worker-1",
        "persona": "backend"
      },
      "created_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

---

### 4.10 `pong`

Heartbeat response.

```json
{
  "type": "pong",
  "payload": {
    "ts": "2024-01-15T12:00:00.000Z"
  }
}
```

---

### 4.11 `error`

Sent in response to a client message that the server could not process.

```json
{
  "type": "error",
  "id": "req_01",
  "payload": {
    "code": "NOT_FOUND",
    "message": "Repository not found"
  }
}
```

---

## 5. Channel Subscriptions

Channels scope which events a client receives. A client must subscribe to a channel before receiving messages from it.

### Channel Types

| Channel Pattern     | Description                                         |
|---------------------|-----------------------------------------------------|
| `repo:{repoId}`     | All repo-level events: task changes, presence, events timeline |
| `agent:{agentId}`   | Agent-specific output stream (high-frequency)       |

**Subscribing to `repo:{repoId}` grants**:
- `task.created`, `task.updated`, `task.deleted`
- `agent.status`
- `presence.snapshot`
- `event.created`

**Subscribing to `agent:{agentId}` grants**:
- `agent.output`

**Authorization**: The server verifies the user has access to the repo before accepting a subscription. Returns `error` with code `FORBIDDEN` if not.

---

## 6. Error Handling

- The server NEVER closes the WebSocket connection due to a recoverable application error — it always sends an `error` message instead.
- The server WILL close the connection (with an appropriate close code) for: authentication failure on upgrade (`4001`), protocol violation (`4002`), or server shutdown (`1001`).
- Clients SHOULD implement exponential backoff reconnection starting at 1s, doubling up to 30s.
- On reconnect, clients MUST re-send all `subscribe` messages and re-send the current `presence.update`.

**WebSocket close codes**:

| Code | Meaning                                      |
|------|----------------------------------------------|
| 1000 | Normal closure                               |
| 1001 | Server going away (restart/shutdown)         |
| 4001 | Authentication failed                        |
| 4002 | Protocol error (malformed message)           |

---

## 7. Heartbeat / Keepalive

- Client SHOULD send a `ping` every **25 seconds**.
- Server responds with `pong` immediately.
- If the server does not receive any message from a client for **60 seconds**, it closes the connection with code `1000`.
- The server also sends a native WebSocket ping frame every **30 seconds**; clients MUST respond with a pong frame (this is handled automatically by most WebSocket libraries).

---

## 8. Message Reference Table

### Client → Server

| Type              | Description                                  |
|-------------------|----------------------------------------------|
| `subscribe`       | Subscribe to a channel                       |
| `unsubscribe`     | Unsubscribe from a channel                   |
| `presence.update` | Announce current navigation/focus            |
| `agent.input`     | Send text input to a running agent           |
| `ping`            | Heartbeat                                    |

### Server → Client

| Type               | Channel        | Description                                  |
|--------------------|----------------|----------------------------------------------|
| `subscribed`       | —              | Subscription accepted                        |
| `unsubscribed`     | —              | Subscription removed                         |
| `agent.output`     | `agent:{id}`   | Streaming output chunk from agent            |
| `agent.status`     | `repo:{id}`    | Agent lifecycle state change                 |
| `task.created`     | `repo:{id}`    | New task created                             |
| `task.updated`     | `repo:{id}`    | Task field(s) changed                        |
| `task.deleted`     | `repo:{id}`    | Task removed                                 |
| `presence.snapshot`| `repo:{id}`    | Full presence state for repo                 |
| `event.created`    | `repo:{id}`    | New timeline event recorded                  |
| `pong`             | —              | Heartbeat response                           |
| `error`            | —              | Error response to a client message           |
