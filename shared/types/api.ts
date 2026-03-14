// TARS v2 — REST API Request / Response Types
// Mirrors the shapes defined in shared/api-contracts.md

import type {
  Repo,
  Task,
  TaskStatus,
  Agent,
  AgentPersona,
  PresenceState,
  TimelineEvent,
  EventType,
  RepoDiff,
  MergeResult,
  User,
  AgentLogsResponse,
} from "./models";

export type { ApiError, ApiErrorResponse } from "./models";

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

export type GetMeResponse = User;

// ---------------------------------------------------------------------------
// Repositories
// ---------------------------------------------------------------------------

export interface ListReposResponse {
  repos: Repo[];
}

export interface CreateRepoRequest {
  name: string;
  github_url: string;
  path: string;
}

export interface UpdateRepoRequest {
  name?: string;
  default_branch?: string;
}

// ---------------------------------------------------------------------------
// Tasks
// ---------------------------------------------------------------------------

export interface ListTasksQuery {
  status?: TaskStatus;
  agent_id?: string;
}

export interface ListTasksResponse {
  tasks: Task[];
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  /** 1–5, default 3 */
  priority?: number;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: number;
  /** Pass `null` to unassign */
  agent_id?: string | null;
}

// ---------------------------------------------------------------------------
// Agents
// ---------------------------------------------------------------------------

export interface ListAgentsQuery {
  status?: Agent["status"];
}

export interface ListAgentsResponse {
  agents: Agent[];
}

export interface SpawnAgentRequest {
  task_id?: string;
  name: string;
  persona?: AgentPersona;
  /** Default: "claude-opus-4-5" */
  model?: string;
  system_prompt?: string;
}

export interface StopAgentResponse {
  id: string;
  status: "stopped";
  stopped_at: string;
}

export interface AgentInputRequest {
  text: string;
}

export interface GetAgentLogsQuery {
  limit?: number;
  offset?: number;
}

export type GetAgentLogsResponse = AgentLogsResponse;

export type MergeStrategy = "merge" | "squash" | "rebase";

export interface MergeAgentBranchRequest {
  target_branch: string;
  strategy?: MergeStrategy;
  commit_message?: string;
}

export type MergeAgentBranchResponse = MergeResult;

// ---------------------------------------------------------------------------
// Presence
// ---------------------------------------------------------------------------

export type GetPresenceResponse = PresenceState;

// ---------------------------------------------------------------------------
// Events / Timeline
// ---------------------------------------------------------------------------

export interface ListEventsQuery {
  limit?: number;
  before?: string;
  after?: string;
  type?: EventType;
  agent_id?: string;
}

export interface ListEventsResponse {
  events: TimelineEvent[];
  has_more: boolean;
}

// ---------------------------------------------------------------------------
// Git Diffs
// ---------------------------------------------------------------------------

export type DiffFormat = "unified" | "stat";

export interface GetDiffQuery {
  base?: string;
  format?: DiffFormat;
}

export type GetDiffResponse = RepoDiff;
