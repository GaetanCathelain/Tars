# TARS v2 — REST API Contracts

> **Contract version**: 1.0.0
> **Base URL**: `http://localhost:8080/api/v1`
> **Auth**: Session cookie `tars_session` (HttpOnly, Secure) set by the OAuth callback. All protected endpoints return `401` if the cookie is absent or invalid.

---

## Table of Contents

1. [Authentication](#1-authentication)
2. [Repositories](#2-repositories)
3. [Tasks](#3-tasks)
4. [Agents](#4-agents)
5. [Presence](#5-presence)
6. [Events / Timeline](#6-events--timeline)
7. [Git Diffs](#7-git-diffs)
8. [Error Schema](#8-error-schema)

---

## 1. Authentication

### 1.1 Initiate GitHub OAuth

```
GET /api/v1/auth/github/login
```

**Auth required**: No
**Query params**: none
**Response**: `302 Found` → redirects to `https://github.com/login/oauth/authorize` with `client_id`, `scope=read:user user:email`, and `state` (CSRF token stored in session).

---

### 1.2 GitHub OAuth Callback

```
GET /api/v1/auth/github/callback
```

**Auth required**: No
**Query params**:

| Name    | Type   | Required | Description             |
|---------|--------|----------|-------------------------|
| `code`  | string | yes      | GitHub authorization code |
| `state` | string | yes      | CSRF state token        |

**Success response**: `302 Found` → redirects to `/dashboard`. Sets `tars_session` cookie.

**Error response**: `302 Found` → redirects to `/login?error=<reason>`.

---

### 1.3 Get Current User

```
GET /api/v1/auth/me
```

**Auth required**: Yes
**Response**: `200 OK`

```json
{
  "id": "usr_01J...",
  "github_id": 1234567,
  "login": "octocat",
  "name": "The Octocat",
  "avatar_url": "https://avatars.githubusercontent.com/u/583231",
  "email": "octocat@github.com",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### 1.4 Logout

```
POST /api/v1/auth/logout
```

**Auth required**: Yes
**Request body**: none
**Response**: `204 No Content`. Clears `tars_session` cookie and invalidates server-side session.

---

## 2. Repositories

Repositories are the Git repos that agents operate on. Each repo maps to a directory on the server and supports worktree-based agent isolation.

### 2.1 List Repositories

```
GET /repos
```

**Auth required**: Yes
**Response**: `200 OK`

```json
{
  "repos": [
    {
      "id": "repo_01J...",
      "name": "my-project",
      "path": "/workspaces/my-project",
      "github_url": "https://github.com/org/my-project",
      "default_branch": "main",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

---

### 2.2 Get Repository

```
GET /repos/:repoId
```

**Auth required**: Yes
**Path params**: `repoId` — repository ID
**Response**: `200 OK` — single repo object (same shape as list item).
**Error**: `404 Not Found` if repo not found.

---

### 2.3 Create Repository

```
POST /repos
```

**Auth required**: Yes
**Request body**:

```json
{
  "name": "my-project",
  "github_url": "https://github.com/org/my-project",
  "path": "/workspaces/my-project"
}
```

| Field        | Type   | Required | Description                                      |
|--------------|--------|----------|--------------------------------------------------|
| `name`       | string | yes      | Human-readable name (unique per user)            |
| `github_url` | string | yes      | Full GitHub HTTPS URL                            |
| `path`       | string | yes      | Absolute path on server where repo is/will be cloned |

**Response**: `201 Created` — full repo object.
**Error**: `409 Conflict` if `name` already exists for this user.

---

### 2.4 Update Repository

```
PATCH /repos/:repoId
```

**Auth required**: Yes
**Request body** (all fields optional):

```json
{
  "name": "renamed-project",
  "default_branch": "develop"
}
```

**Response**: `200 OK` — updated repo object.

---

### 2.5 Delete Repository

```
DELETE /repos/:repoId
```

**Auth required**: Yes
**Response**: `204 No Content`.
**Note**: Does not delete files from disk. Removes DB record and all associated agents/tasks.

---

## 3. Tasks

Tasks represent discrete units of work assigned to agents. They map to a Kanban-style board in the UI.

### 3.1 List Tasks

```
GET /repos/:repoId/tasks
```

**Auth required**: Yes
**Query params**:

| Name     | Type   | Required | Description                                      |
|----------|--------|----------|--------------------------------------------------|
| `status` | string | no       | Filter: `pending`, `in_progress`, `done`, `cancelled` |
| `agent_id` | string | no     | Filter by assigned agent                         |

**Response**: `200 OK`

```json
{
  "tasks": [
    {
      "id": "task_01J...",
      "repo_id": "repo_01J...",
      "title": "Implement OAuth login",
      "description": "Add GitHub OAuth 2.0 flow to the backend.",
      "status": "in_progress",
      "priority": 2,
      "agent_id": "agent_01J...",
      "created_by": "usr_01J...",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

---

### 3.2 Get Task

```
GET /repos/:repoId/tasks/:taskId
```

**Auth required**: Yes
**Response**: `200 OK` — single task object.

---

### 3.3 Create Task

```
POST /repos/:repoId/tasks
```

**Auth required**: Yes
**Request body**:

```json
{
  "title": "Implement OAuth login",
  "description": "Add GitHub OAuth 2.0 flow to the backend.",
  "priority": 2
}
```

| Field         | Type    | Required | Description                            |
|---------------|---------|----------|----------------------------------------|
| `title`       | string  | yes      | Short task title (max 255 chars)       |
| `description` | string  | no       | Full markdown description              |
| `priority`    | integer | no       | 1 (highest) – 5 (lowest), default: 3  |

**Response**: `201 Created` — full task object.

---

### 3.4 Update Task

```
PATCH /repos/:repoId/tasks/:taskId
```

**Auth required**: Yes
**Request body** (all fields optional):

```json
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "done",
  "priority": 1,
  "agent_id": "agent_01J..."
}
```

| Field       | Type    | Required | Allowed values                                    |
|-------------|---------|----------|---------------------------------------------------|
| `status`    | string  | no       | `pending`, `in_progress`, `done`, `cancelled`     |
| `agent_id`  | string  | no       | ID of agent to assign, or `null` to unassign      |

**Response**: `200 OK` — updated task object.
**Side effect**: Emits a `task_updated` WebSocket event to all connected clients in the repo channel.

---

### 3.5 Delete Task

```
DELETE /repos/:repoId/tasks/:taskId
```

**Auth required**: Yes
**Response**: `204 No Content`.

---

## 4. Agents

Agents are Claude Code CLI processes running inside Git worktrees. Each agent is scoped to a repo and optionally to a task.

### 4.1 List Agents

```
GET /repos/:repoId/agents
```

**Auth required**: Yes
**Query params**:

| Name     | Type   | Required | Description                                          |
|----------|--------|----------|------------------------------------------------------|
| `status` | string | no       | Filter: `starting`, `running`, `stopped`, `crashed`  |

**Response**: `200 OK`

```json
{
  "agents": [
    {
      "id": "agent_01J...",
      "repo_id": "repo_01J...",
      "task_id": "task_01J...",
      "name": "claude-worker-1",
      "persona": "backend",
      "status": "running",
      "worktree_path": "/workspaces/my-project/.tars/agents/agent_01J...",
      "branch": "tars/agent-01J...",
      "pid": 12345,
      "started_at": "2024-01-15T10:30:00Z",
      "stopped_at": null
    }
  ]
}
```

---

### 4.2 Get Agent

```
GET /repos/:repoId/agents/:agentId
```

**Auth required**: Yes
**Response**: `200 OK` — single agent object.

---

### 4.3 Spawn Agent

```
POST /repos/:repoId/agents
```

**Auth required**: Yes
**Request body**:

```json
{
  "task_id": "task_01J...",
  "name": "claude-worker-1",
  "persona": "backend",
  "model": "claude-opus-4-5",
  "system_prompt": "You are a senior Go backend engineer..."
}
```

| Field           | Type   | Required | Description                                                     |
|-----------------|--------|----------|-----------------------------------------------------------------|
| `task_id`       | string | no       | Task to link this agent to                                      |
| `name`          | string | yes      | Human-readable agent name (unique per repo)                     |
| `persona`       | string | no       | Persona tag: `backend`, `frontend`, `devops`, `qa`, `general`  |
| `model`         | string | no       | Claude model ID, default: `claude-opus-4-5`                    |
| `system_prompt` | string | no       | Override system prompt                                          |

**Response**: `201 Created` — full agent object.
**Side effect**: Creates Git worktree, starts `claude` process, begins streaming output over WebSocket channel `agent:{agentId}`.

---

### 4.4 Stop Agent

```
POST /repos/:repoId/agents/:agentId/stop
```

**Auth required**: Yes
**Request body**: none
**Response**: `200 OK`

```json
{
  "id": "agent_01J...",
  "status": "stopped",
  "stopped_at": "2024-01-15T12:00:00Z"
}
```

**Side effect**: Sends SIGTERM to the Claude process. If not exited within 5s, sends SIGKILL.

---

### 4.5 Send Input to Agent

```
POST /repos/:repoId/agents/:agentId/input
```

**Auth required**: Yes
**Request body**:

```json
{
  "text": "Please focus on the authentication module first."
}
```

| Field  | Type   | Required | Description                     |
|--------|--------|----------|---------------------------------|
| `text` | string | yes      | Text to write to agent's stdin  |

**Response**: `204 No Content`.

---

### 4.6 Get Agent Logs

```
GET /repos/:repoId/agents/:agentId/logs
```

**Auth required**: Yes
**Query params**:

| Name     | Type    | Required | Description                           |
|----------|---------|----------|---------------------------------------|
| `limit`  | integer | no       | Max lines, default 500, max 5000      |
| `offset` | integer | no       | Line offset for pagination, default 0 |

**Response**: `200 OK`

```json
{
  "agent_id": "agent_01J...",
  "lines": [
    {
      "seq": 1,
      "ts": "2024-01-15T10:30:01Z",
      "stream": "stdout",
      "text": "Claude Code starting..."
    }
  ],
  "total": 1234
}
```

---

### 4.7 Merge Agent Branch

```
POST /repos/:repoId/agents/:agentId/merge
```

**Auth required**: Yes
**Request body**:

```json
{
  "target_branch": "main",
  "strategy": "squash",
  "commit_message": "feat: implement OAuth login\n\nCo-authored-by: TARS Agent claude-worker-1"
}
```

| Field            | Type   | Required | Allowed values                       |
|------------------|--------|----------|--------------------------------------|
| `target_branch`  | string | yes      | Branch to merge the agent branch into|
| `strategy`       | string | no       | `merge`, `squash`, `rebase` — default `squash` |
| `commit_message` | string | no       | Override commit/squash message       |

**Response**: `200 OK`

```json
{
  "merged": true,
  "target_branch": "main",
  "agent_branch": "tars/agent-01J...",
  "commit_sha": "abc1234..."
}
```

**Error**: `409 Conflict` if merge conflicts exist — includes conflict details.

---

## 5. Presence

Tracks which users are currently online and which agents they are viewing.

### 5.1 Get Presence

```
GET /repos/:repoId/presence
```

**Auth required**: Yes
**Response**: `200 OK`

```json
{
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
```

---

## 6. Events / Timeline

Events are an immutable audit log of everything that happens in a repo (agent spawned, task updated, merge completed, etc.).

### 6.1 List Events

```
GET /repos/:repoId/events
```

**Auth required**: Yes
**Query params**:

| Name       | Type    | Required | Description                                 |
|------------|---------|----------|---------------------------------------------|
| `limit`    | integer | no       | Max events returned, default 50, max 200    |
| `before`   | string  | no       | ISO 8601 timestamp — return events before this time |
| `after`    | string  | no       | ISO 8601 timestamp — return events after this time  |
| `type`     | string  | no       | Filter by event type (see Event Types below)|
| `agent_id` | string  | no       | Filter by agent                             |

**Response**: `200 OK`

```json
{
  "events": [
    {
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
  ],
  "has_more": true
}
```

**Event types**:

| Type                | Description                              |
|---------------------|------------------------------------------|
| `agent.spawned`     | An agent was started                     |
| `agent.stopped`     | An agent was stopped by user             |
| `agent.crashed`     | An agent exited unexpectedly             |
| `agent.merged`      | An agent's branch was merged             |
| `task.created`      | A new task was created                   |
| `task.updated`      | A task's fields were updated             |
| `task.deleted`      | A task was deleted                       |
| `task.assigned`     | A task was assigned to an agent          |
| `repo.created`      | A repository was registered              |
| `user.joined`       | A user connected (presence)              |
| `user.left`         | A user disconnected (presence)           |

---

## 7. Git Diffs

### 7.1 Get Agent Worktree Diff

```
GET /repos/:repoId/agents/:agentId/diff
```

**Auth required**: Yes
**Query params**:

| Name     | Type   | Required | Description                                    |
|----------|--------|----------|------------------------------------------------|
| `base`   | string | no       | Base ref to diff against, default: repo's `default_branch` |
| `format` | string | no       | `unified` (default), `stat`                    |

**Response**: `200 OK`

```json
{
  "agent_id": "agent_01J...",
  "base_ref": "main",
  "head_ref": "tars/agent-01J...",
  "stats": {
    "files_changed": 3,
    "insertions": 120,
    "deletions": 45
  },
  "files": [
    {
      "path": "backend/internal/auth/github.go",
      "status": "modified",
      "additions": 80,
      "deletions": 20,
      "patch": "@@ -1,5 +1,8 @@\n ..."
    },
    {
      "path": "backend/internal/auth/session.go",
      "status": "added",
      "additions": 40,
      "deletions": 0,
      "patch": "@@ -0,0 +1,40 @@\n ..."
    }
  ]
}
```

**File status values**: `added`, `modified`, `deleted`, `renamed`, `copied`

---

## 8. Error Schema

All error responses share a common JSON body:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Repository not found",
    "details": {}
  }
}
```

| Field     | Type   | Description                             |
|-----------|--------|-----------------------------------------|
| `code`    | string | Machine-readable error code (SCREAMING_SNAKE_CASE) |
| `message` | string | Human-readable description              |
| `details` | object | Optional extra context (e.g., validation errors) |

**Standard HTTP status codes**:

| Status | When used                                              |
|--------|--------------------------------------------------------|
| 200    | Success with body                                      |
| 201    | Resource created                                       |
| 204    | Success, no body                                       |
| 302    | Redirect (OAuth only)                                  |
| 400    | Bad request / validation failure                       |
| 401    | Not authenticated                                      |
| 403    | Authenticated but not authorized                       |
| 404    | Resource not found                                     |
| 409    | Conflict (duplicate name, merge conflict)              |
| 500    | Internal server error                                  |

**Common error codes**:

| Code                  | HTTP | Description                          |
|-----------------------|------|--------------------------------------|
| `UNAUTHORIZED`        | 401  | Missing or invalid session            |
| `FORBIDDEN`           | 403  | Access denied to resource             |
| `NOT_FOUND`           | 404  | Resource does not exist               |
| `CONFLICT`            | 409  | Duplicate or conflicting resource     |
| `VALIDATION_ERROR`    | 400  | Request body failed validation        |
| `AGENT_NOT_RUNNING`   | 400  | Operation requires a running agent    |
| `MERGE_CONFLICT`      | 409  | Git merge has unresolvable conflicts  |
| `INTERNAL_ERROR`      | 500  | Unexpected server error               |

---

## Appendix: ID Format

All resource IDs use the ULID format (`01J...`) — lexicographically sortable, URL-safe, 26 characters. Generated server-side.

## Appendix: Pagination

Endpoints returning lists use cursor-based pagination via `before` / `after` timestamps where applicable, or `limit`/`offset` for simpler lists. The `has_more` boolean indicates additional pages exist.
