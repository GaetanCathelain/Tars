import type { Task, Agent, TimelineEvent, PresenceUser } from '$shared/types/models';

// ---------------------------------------------------------------------------
// Envelope
// ---------------------------------------------------------------------------

export interface Envelope {
	type: string;
	id?: string;
	channel?: string;
	payload: object;
}

// ---------------------------------------------------------------------------
// Client → Server message payloads
// ---------------------------------------------------------------------------

export interface SubscribePayload {
	channel: string;
}

export interface UnsubscribePayload {
	channel: string;
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
// Server → Client message payloads
// ---------------------------------------------------------------------------

export interface AgentOutputPayload {
	agent_id: string;
	seq: number;
	ts: string;
	stream: 'stdout' | 'stderr';
	text: string;
}

export interface AgentStatusPayload {
	agent_id: string;
	status: 'starting' | 'running' | 'stopped' | 'crashed';
	exit_code: number | null;
	ts: string;
}

export interface TaskUpdatedPayload {
	task: Task;
}

export interface TaskCreatedPayload {
	task: Task;
}

export interface TaskDeletedPayload {
	task_id: string;
}

export interface PresenceSnapshotPayload {
	repo_id: string;
	users: PresenceUser[];
}

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
// Typed inbound message union
// ---------------------------------------------------------------------------

export type InboundMessage =
	| { type: 'subscribed'; id?: string; channel?: string; payload: SubscribePayload }
	| { type: 'unsubscribed'; id?: string; channel?: string; payload: UnsubscribePayload }
	| { type: 'agent.output'; channel?: string; payload: AgentOutputPayload }
	| { type: 'agent.status'; channel?: string; payload: AgentStatusPayload }
	| { type: 'task.created'; channel?: string; payload: TaskCreatedPayload }
	| { type: 'task.updated'; channel?: string; payload: TaskUpdatedPayload }
	| { type: 'task.deleted'; channel?: string; payload: TaskDeletedPayload }
	| { type: 'presence.snapshot'; channel?: string; payload: PresenceSnapshotPayload }
	| { type: 'event.created'; channel?: string; payload: EventCreatedPayload }
	| { type: 'pong'; payload: PongPayload }
	| { type: 'error'; id?: string; payload: WsErrorPayload };
