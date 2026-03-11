import { authStore } from './auth.svelte';
import { messagesStore } from './messages.svelte';
import { workersStore } from './workers.svelte';
import { tasksStore } from './tasks.svelte';
import type { Message, WorkerSession } from '../types';

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected';

type WorkerOutputCallback = (sessionId: string, data: Uint8Array) => void;

function createWebSocketStore() {
	let status = $state<ConnectionStatus>('disconnected');
	let ws = $state<WebSocket | null>(null);
	let reconnectAttempts = $state(0);
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	let subscribedTaskId = $state<string | null>(null);

	const outputListeners = new Map<string, Set<WorkerOutputCallback>>();

	function getWsUrl(): string {
		if (typeof window === 'undefined') return '';
		const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		return `${proto}//${window.location.host}/ws`;
	}

	function connect() {
		if (!authStore.token || typeof window === 'undefined') return;
		if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) return;

		status = 'connecting';
		const url = `${getWsUrl()}?token=${encodeURIComponent(authStore.token)}`;
		const socket = new WebSocket(url);

		socket.onopen = () => {
			status = 'connected';
			reconnectAttempts = 0;
			ws = socket;

			// Re-subscribe to current task if any
			if (subscribedTaskId) {
				send({ type: 'subscribe', task_id: subscribedTaskId });
			}
		};

		socket.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				handleEvent(data);
			} catch (e) {
				console.error('[ws] Failed to parse message:', e);
			}
		};

		socket.onclose = () => {
			ws = null;
			status = 'disconnected';
			scheduleReconnect();
		};

		socket.onerror = () => {
			// onclose will fire after onerror
		};
	}

	function handleEvent(data: Record<string, unknown>) {
		switch (data.type) {
			case 'message': {
				const msg = data.message as Message;
				if (msg && msg.task_id === subscribedTaskId) {
					messagesStore.addMessage(msg);
				}
				break;
			}
			case 'worker_start': {
				const session = data.session as WorkerSession;
				if (session) {
					workersStore.onWorkerStart(session);
					if (data.task_id === subscribedTaskId) {
						messagesStore.addWorkerEvent(session.task_id, session.id, 'start');
					}
				}
				break;
			}
			case 'worker_output': {
				const sessionId = data.session_id as string;
				const b64 = data.data as string;
				if (sessionId && b64) {
					const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
					// Dispatch to registered terminal listeners
					const listeners = outputListeners.get(sessionId);
					if (listeners) {
						for (const cb of listeners) {
							cb(sessionId, bytes);
						}
					}
				}
				break;
			}
			case 'worker_end': {
				const sessionId = data.session_id as string;
				const exitCode = data.exit_code as number | undefined;
				if (sessionId) {
					workersStore.onWorkerEnd(sessionId, exitCode ?? 0);
				}
				break;
			}
			case 'task_status': {
				const taskId = data.task_id as string;
				const taskStatus = data.status as string;
				if (taskId && taskStatus) {
					tasksStore.updateTaskStatus(taskId, taskStatus as 'open' | 'running' | 'completed' | 'failed');
				}
				break;
			}
		}
	}

	function send(msg: Record<string, unknown>) {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify(msg));
		}
	}

	function scheduleReconnect() {
		if (reconnectTimer) clearTimeout(reconnectTimer);
		if (!authStore.token) return;

		const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
		reconnectTimer = setTimeout(() => {
			reconnectAttempts++;
			connect();
		}, delay);
	}

	function subscribeToTask(taskId: string) {
		if (subscribedTaskId && subscribedTaskId !== taskId) {
			send({ type: 'unsubscribe', task_id: subscribedTaskId });
		}
		subscribedTaskId = taskId;
		send({ type: 'subscribe', task_id: taskId });
	}

	function unsubscribeFromTask(taskId: string) {
		if (subscribedTaskId === taskId) {
			send({ type: 'unsubscribe', task_id: taskId });
			subscribedTaskId = null;
		}
	}

	function onWorkerOutput(sessionId: string, callback: WorkerOutputCallback) {
		if (!outputListeners.has(sessionId)) {
			outputListeners.set(sessionId, new Set());
		}
		outputListeners.get(sessionId)!.add(callback);

		// Return unsubscribe function
		return () => {
			const set = outputListeners.get(sessionId);
			if (set) {
				set.delete(callback);
				if (set.size === 0) outputListeners.delete(sessionId);
			}
		};
	}

	function disconnect() {
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
			reconnectTimer = null;
		}
		if (ws) {
			ws.close();
			ws = null;
		}
		status = 'disconnected';
		subscribedTaskId = null;
	}

	return {
		get status() { return status; },
		get subscribedTaskId() { return subscribedTaskId; },
		connect,
		disconnect,
		subscribeToTask,
		unsubscribeFromTask,
		onWorkerOutput,
		send
	};
}

export const wsStore = createWebSocketStore();
