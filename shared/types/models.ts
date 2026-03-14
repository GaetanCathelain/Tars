// TARS v2 — Core Domain Models
// These represent the server-side resource shapes returned by the REST API.

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

export interface User {
  id: string;
  github_id: number;
  login: string;
  name: string;
  avatar_url: string;
  email: string;
  created_at: string; // ISO 8601
}

// ---------------------------------------------------------------------------
// Repository
// ---------------------------------------------------------------------------

export interface Repo {
  id: string;
  name: string;
  path: string;
  github_url: string;
  default_branch: string;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// Task
// ---------------------------------------------------------------------------

export type TaskStatus = "pending" | "in_progress" | "done" | "cancelled";

export interface Task {
  id: string;
  repo_id: string;
  title: string;
  description: string;
  status: TaskStatus;
  /** 1 (highest) – 5 (lowest) */
  priority: number;
  agent_id: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

// ---------------------------------------------------------------------------
// Agent
// ---------------------------------------------------------------------------

export type AgentStatus = "starting" | "running" | "stopped" | "crashed";

export type AgentPersona =
  | "backend"
  | "frontend"
  | "devops"
  | "qa"
  | "general";

export interface Agent {
  id: string;
  repo_id: string;
  task_id: string | null;
  name: string;
  persona: AgentPersona;
  status: AgentStatus;
  worktree_path: string;
  branch: string;
  pid: number | null;
  started_at: string;
  stopped_at: string | null;
}

// ---------------------------------------------------------------------------
// Agent Log Line
// ---------------------------------------------------------------------------

export type LogStream = "stdout" | "stderr";

export interface AgentLogLine {
  seq: number;
  ts: string;
  stream: LogStream;
  text: string;
}

export interface AgentLogsResponse {
  agent_id: string;
  lines: AgentLogLine[];
  total: number;
}

// ---------------------------------------------------------------------------
// Presence
// ---------------------------------------------------------------------------

export interface PresenceUser {
  user_id: string;
  login: string;
  avatar_url: string;
  viewing_agent_id: string | null;
  last_seen: string;
}

export interface PresenceState {
  repo_id: string;
  users: PresenceUser[];
}

// ---------------------------------------------------------------------------
// Event / Timeline
// ---------------------------------------------------------------------------

export type EventType =
  | "agent.spawned"
  | "agent.stopped"
  | "agent.crashed"
  | "agent.merged"
  | "task.created"
  | "task.updated"
  | "task.deleted"
  | "task.assigned"
  | "repo.created"
  | "user.joined"
  | "user.left";

export type ActorType = "user" | "agent" | "system";

export interface TimelineEvent {
  id: string;
  repo_id: string;
  type: EventType;
  actor_type: ActorType;
  actor_id: string | null;
  agent_id: string | null;
  task_id: string | null;
  payload: Record<string, unknown>;
  created_at: string;
}

// ---------------------------------------------------------------------------
// Git Diff
// ---------------------------------------------------------------------------

export type FileStatus =
  | "added"
  | "modified"
  | "deleted"
  | "renamed"
  | "copied";

export interface FileDiff {
  path: string;
  status: FileStatus;
  additions: number;
  deletions: number;
  patch: string;
}

export interface DiffStats {
  files_changed: number;
  insertions: number;
  deletions: number;
}

export interface RepoDiff {
  agent_id: string;
  base_ref: string;
  head_ref: string;
  stats: DiffStats;
  files: FileDiff[];
}

// ---------------------------------------------------------------------------
// Merge Result
// ---------------------------------------------------------------------------

export interface MergeResult {
  merged: boolean;
  target_branch: string;
  agent_branch: string;
  commit_sha: string;
}

// ---------------------------------------------------------------------------
// Error
// ---------------------------------------------------------------------------

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiErrorResponse {
  error: ApiError;
}
