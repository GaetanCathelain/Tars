import { wsClient } from './client.svelte';
import type { InboundMessage } from './types';
import { repos } from '$lib/stores/repos.svelte';
import { tasks } from '$lib/stores/tasks.svelte';
import { agents } from '$lib/stores/agents.svelte';
import { presence } from '$lib/stores/presence.svelte';
import { events } from '$lib/stores/events.svelte';

let unregister: (() => void) | null = null;

function handleMessage(msg: InboundMessage): void {
	switch (msg.type) {
		case 'task.created':
			tasks.addTask(msg.payload.task);
			break;

		case 'task.updated':
			tasks.updateTask(msg.payload.task);
			break;

		case 'task.deleted':
			tasks.removeTask(msg.payload.task_id);
			break;

		case 'agent.status':
			agents.updateStatus(msg.payload.agent_id, msg.payload.status, msg.payload.exit_code);
			break;

		case 'agent.output':
			agents.appendOutput(msg.payload.agent_id, {
				seq: msg.payload.seq,
				ts: msg.payload.ts,
				stream: msg.payload.stream,
				text: msg.payload.text
			});
			break;

		case 'presence.snapshot':
			presence.setSnapshot(msg.payload.repo_id, msg.payload.users);
			break;

		case 'event.created':
			events.addEvent(msg.payload.event);
			break;

		case 'pong':
			// No-op — heartbeat confirmed
			break;

		case 'error':
			console.warn('[ws] server error:', msg.payload.code, msg.payload.message);
			break;

		default:
			// Unknown message type — ignore
			break;
	}
}

/** Register the router. Call once at app startup. */
export function startRouter(): void {
	if (unregister) return;
	unregister = wsClient.onMessage(handleMessage);
}

/** Deregister the router. */
export function stopRouter(): void {
	unregister?.();
	unregister = null;
}
