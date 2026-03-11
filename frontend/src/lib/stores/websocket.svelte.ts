const MOCK_MODE = true;

type ConnectionStatus = 'connected' | 'disconnected' | 'connecting';

function createWebSocketStore() {
	let status = $state<ConnectionStatus>('disconnected');
	let ws = $state<WebSocket | null>(null);
	let subscribedTaskId = $state<string | null>(null);

	function connect(token: string) {
		if (MOCK_MODE) {
			status = 'disconnected';
			return;
		}

		status = 'connecting';
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${protocol}//${window.location.host}/api/ws?token=${token}`;

		const socket = new WebSocket(wsUrl);

		socket.onopen = () => {
			status = 'connected';
			ws = socket;
		};

		socket.onclose = () => {
			status = 'disconnected';
			ws = null;
		};

		socket.onerror = () => {
			status = 'disconnected';
		};
	}

	function disconnect() {
		ws?.close();
		ws = null;
		status = 'disconnected';
	}

	function subscribeToTask(taskId: string) {
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
		unsubscribeFromTask
	};
}

export const wsStore = createWebSocketStore();
