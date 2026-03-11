import { workersStore } from '$lib/stores/workers.svelte';
import { tasksStore } from '$lib/stores/tasks.svelte';
import { messagesStore } from '$lib/stores/messages.svelte';

type ConnectionStatus = 'connected' | 'disconnected' | 'connecting';
type OutputCallback = (data: Uint8Array) => void;

function createWebSocketStore() {
	let status = $state<ConnectionStatus>('disconnected');
	let ws = $state<WebSocket | null>(null);
	let subscribedTaskId = $state<string | null>(null);
	let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

	// Worker output subscription system
	const outputSubscribers = new Map<string, Set<OutputCallback>>();

	function subscribe(sessionId: string, callback: OutputCallback): void {
		if (!outputSubscribers.has(sessionId)) {
			outputSubscribers.set(sessionId, new Set());
		}
		outputSubscribers.get(sessionId)!.add(callback);
	}

	function unsubscribe(sessionId: string, callback: OutputCallback): void {
		const subs = outputSubscribers.get(sessionId);
		if (subs) {
			subs.delete(callback);
			if (subs.size === 0) outputSubscribers.delete(sessionId);
		}
	}

	function dispatchOutput(sessionId: string, data: Uint8Array): void {
		const subs = outputSubscribers.get(sessionId);
		if (subs) {
			for (const cb of subs) {
				cb(data);
			}
		}
	}

	function handleMessage(event: MessageEvent): void {
		try {
			const msg = JSON.parse(event.data);
			switch (msg.type) {
				case 'worker_output': {
					const decoded = atob(msg.data);
					const bytes = Uint8Array.from(decoded, (c) => c.charCodeAt(0));
					dispatchOutput(msg.session_id, bytes);
					break;
				}
				case 'worker_start': {
					if (msg.session) {
						workersStore.addWorker(msg.session);
					}
					break;
				}
				case 'worker_end': {
					const exitStatus = msg.exit_code === 0 ? 'completed' : 'failed';
					workersStore.updateWorkerStatus(msg.session_id, exitStatus, msg.exit_code);
					break;
				}
				case 'task_status': {
					tasksStore.updateTaskStatus(msg.task_id, msg.status);
					break;
				}
				case 'message': {
					if (msg.message) {
						messagesStore.addMessage(msg.message);
					}
					break;
				}
			}
		} catch {
			// Ignore non-JSON messages
		}
	}

	function connect(token: string) {
		if (ws && ws.readyState !== WebSocket.CLOSED) return;

		status = 'connecting';
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${protocol}//${window.location.host}/ws?token=${token}`;

		const socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			status = 'connected';
			ws = socket;
			// Re-subscribe to task if we had one
			if (subscribedTaskId) {
				socket.send(JSON.stringify({ type: 'subscribe', task_id: subscribedTaskId }));
			}
		};

		socket.onmessage = handleMessage;

		socket.onclose = () => {
			status = 'disconnected';
			ws = null;
			// Auto-reconnect after 3 seconds
			if (reconnectTimer) clearTimeout(reconnectTimer);
			reconnectTimer = setTimeout(() => {
				if (token) connect(token);
			}, 3000);
		};

		socket.onerror = () => {
			status = 'disconnected';
		};
	}

	function disconnect() {
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
			reconnectTimer = null;
		}
		ws?.close();
		ws = null;
		status = 'disconnected';
	}

	function subscribeToTask(taskId: string) {
		// Unsubscribe from previous task first
		if (subscribedTaskId && subscribedTaskId !== taskId) {
			unsubscribeFromTask();
		}
		subscribedTaskId = taskId;
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({ type: 'subscribe', task_id: taskId }));
		}
	}

	function unsubscribeFromTask() {
		if (ws && ws.readyState === WebSocket.OPEN && subscribedTaskId) {
			ws.send(JSON.stringify({ type: 'unsubscribe', task_id: subscribedTaskId }));
		}
		subscribedTaskId = null;
	}

	return {
		get status() { return status; },
		get subscribedTaskId() { return subscribedTaskId; },
		connect,
		disconnect,
		subscribeToTask,
		unsubscribeFromTask,
		subscribe,
		unsubscribe
	};
}

export const wsStore = createWebSocketStore();
