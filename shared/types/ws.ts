// TARS v2 — WebSocket Message Types
// Mirrors the shapes defined in shared/ws-protocol.md

import type { Task, Agent, AgentStatus, LogStream, PresenceState, TimelineEvent } from "./models";

// ---------------------------------------------------------------------------
// Base Envelope
// ---------------------------------------------------------------------------

export interface WsEnvelope<T extends string, P extends object> {
  type: T;
  /** Client-generated request ID for correlation */
  id?: string;
  /** Channel this message belongs to */
  channel?: string;
  payload: P;
}

// ---------------------------------------------------------------------------
// Channel Names
// ---------------------------------------------------------------------------

/** `repo:{repoId}` */
export type RepoChannel = `repo:${string}`;
/** `agent:{agentId}` */
export type AgentChannel = `agent:${string}`;
export type Channel = RepoChannel | AgentChannel;

// ---------------------------------------------------------------------------
// Client → Server Message Payloads
// ---------------------------------------------------------------------------

export interface SubscribePayload {
  channel: Channel;
}

export interface UnsubscribePayload {
  channel: Channel;
}

export interface PresenceUpdatePayload {
  repo_id: string;
  viewing_agent_id?: string | null;
}

export interface AgentInputPayload {
  agent_id: string;
  text: string;
}

// ---------------------------------------------------------------------------
// Client → Server Messages (discriminated union)
// ---------------------------------------------------------------------------

export type WsSubscribeMessage = WsEnvelope<"subscribe", SubscribePayload>;
export type WsUnsubscribeMessage = WsEnvelope<"unsubscribe", UnsubscribePayload>;
export type WsPresenceUpdateMessage = WsEnvelope<"presence.update", PresenceUpdatePayload>;
export type WsAgentInputMessage = WsEnvelope<"agent.input", AgentInputPayload>;
export type WsPingMessage = WsEnvelope<"ping", Record<string, never>>;

export type ClientMessage =
  | WsSubscribeMessage
  | WsUnsubscribeMessage
  | WsPresenceUpdateMessage
  | WsAgentInputMessage
  | WsPingMessage;

// ---------------------------------------------------------------------------
// Server → Client Message Payloads
// ---------------------------------------------------------------------------

export interface SubscribedPayload {
  channel: Channel;
}

export interface UnsubscribedPayload {
  channel: Channel;
}

export interface AgentOutputPayload {
  agent_id: string;
  seq: number;
  ts: string;
  stream: LogStream;
  text: string;
}

export interface AgentStatusPayload {
  agent_id: string;
  status: AgentStatus;
  exit_code: number | null;
  ts: string;
}

export interface TaskCreatedPayload {
  task: Task;
}

export interface TaskUpdatedPayload {
  task: Task;
}

export interface TaskDeletedPayload {
  task_id: string;
}

export type PresenceSnapshotPayload = PresenceState;

export interface EventCreatedPayload {
  event: TimelineEvent;
}

export interface PongPayload {
  ts: string;
}

export interface WsErrorPayload {
  code: string;
  message: string;
}

// ---------------------------------------------------------------------------
// Server → Client Messages (discriminated union)
// ---------------------------------------------------------------------------

export type WsSubscribedMessage = WsEnvelope<"subscribed", SubscribedPayload>;
export type WsUnsubscribedMessage = WsEnvelope<"unsubscribed", UnsubscribedPayload>;
export type WsAgentOutputMessage = WsEnvelope<"agent.output", AgentOutputPayload>;
export type WsAgentStatusMessage = WsEnvelope<"agent.status", AgentStatusPayload>;
export type WsTaskCreatedMessage = WsEnvelope<"task.created", TaskCreatedPayload>;
export type WsTaskUpdatedMessage = WsEnvelope<"task.updated", TaskUpdatedPayload>;
export type WsTaskDeletedMessage = WsEnvelope<"task.deleted", TaskDeletedPayload>;
export type WsPresenceSnapshotMessage = WsEnvelope<"presence.snapshot", PresenceSnapshotPayload>;
export type WsEventCreatedMessage = WsEnvelope<"event.created", EventCreatedPayload>;
export type WsPongMessage = WsEnvelope<"pong", PongPayload>;
export type WsErrorMessage = WsEnvelope<"error", WsErrorPayload>;

export type ServerMessage =
  | WsSubscribedMessage
  | WsUnsubscribedMessage
  | WsAgentOutputMessage
  | WsAgentStatusMessage
  | WsTaskCreatedMessage
  | WsTaskUpdatedMessage
  | WsTaskDeletedMessage
  | WsPresenceSnapshotMessage
  | WsEventCreatedMessage
  | WsPongMessage
  | WsErrorMessage;

// ---------------------------------------------------------------------------
// Type guard helpers
// ---------------------------------------------------------------------------

export function isAgentOutputMessage(msg: ServerMessage): msg is WsAgentOutputMessage {
  return msg.type === "agent.output";
}

export function isAgentStatusMessage(msg: ServerMessage): msg is WsAgentStatusMessage {
  return msg.type === "agent.status";
}

export function isTaskCreatedMessage(msg: ServerMessage): msg is WsTaskCreatedMessage {
  return msg.type === "task.created";
}

export function isTaskUpdatedMessage(msg: ServerMessage): msg is WsTaskUpdatedMessage {
  return msg.type === "task.updated";
}

export function isTaskDeletedMessage(msg: ServerMessage): msg is WsTaskDeletedMessage {
  return msg.type === "task.deleted";
}

export function isPresenceSnapshotMessage(msg: ServerMessage): msg is WsPresenceSnapshotMessage {
  return msg.type === "presence.snapshot";
}

export function isEventCreatedMessage(msg: ServerMessage): msg is WsEventCreatedMessage {
  return msg.type === "event.created";
}

export function isErrorMessage(msg: ServerMessage): msg is WsErrorMessage {
  return msg.type === "error";
}
