import type { Message } from '../types';

const useMockData = true;

const mockMessages: Record<string, Message[]> = {
	'task-1': [
		{
			id: 'msg-1',
			task_id: 'task-1',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Set up JWT authentication with bcrypt password hashing. Use SQLite for storage.',
			created_at: '2026-03-10T14:00:30Z'
		},
		{
			id: 'msg-2',
			task_id: 'task-1',
			sender_type: 'tars',
			content: 'Starting authentication implementation. Creating auth middleware, login/register endpoints, and JWT token generation.',
			created_at: '2026-03-10T14:01:00Z'
		},
		{
			id: 'msg-3',
			task_id: 'task-1',
			sender_type: 'system',
			content: 'Worker session started: claude-code-auth-impl',
			created_at: '2026-03-10T14:01:05Z'
		},
		{
			id: 'msg-4',
			task_id: 'task-1',
			sender_type: 'tars',
			content: 'Authentication system complete. Created:\n- POST /api/auth/register\n- POST /api/auth/login\n- JWT middleware with 24h expiry\n- bcrypt password hashing (cost 12)\n- SQLite user table with unique username constraint',
			created_at: '2026-03-10T15:28:00Z'
		},
		{
			id: 'msg-5',
			task_id: 'task-1',
			sender_type: 'system',
			content: 'Worker session completed (exit code 0)',
			created_at: '2026-03-10T15:30:00Z'
		}
	],
	'task-2': [
		{
			id: 'msg-6',
			task_id: 'task-2',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Build CRUD endpoints for tasks and messages. Include pagination for message listing.',
			created_at: '2026-03-10T16:00:30Z'
		},
		{
			id: 'msg-7',
			task_id: 'task-2',
			sender_type: 'tars',
			content: 'Working on REST API. Creating task and message models, handlers, and route registration.',
			created_at: '2026-03-10T16:01:00Z'
		},
		{
			id: 'msg-8',
			task_id: 'task-2',
			sender_type: 'system',
			content: 'Worker session started: claude-code-api-endpoints',
			created_at: '2026-03-10T16:01:05Z'
		},
		{
			id: 'msg-9',
			task_id: 'task-2',
			sender_type: 'tars',
			content: 'Task endpoints done. Working on message endpoints with cursor-based pagination now.',
			created_at: '2026-03-11T00:10:00Z'
		}
	],
	'task-3': [
		{
			id: 'msg-10',
			task_id: 'task-3',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Add WebSocket support for real-time message streaming from worker sessions.',
			created_at: '2026-03-11T00:30:30Z'
		}
	]
};

function createMessagesStore() {
	let messages = $state<Message[]>([]);
	let currentTaskId = $state<string | null>(null);
	let loading = $state(false);

	async function fetchMessages(taskId: string) {
		currentTaskId = taskId;
		loading = true;
		try {
			if (useMockData) {
				await new Promise((r) => setTimeout(r, 150));
				messages = mockMessages[taskId] ? [...mockMessages[taskId]] : [];
				return;
			}

			const { api } = await import('../api');
			messages = await api.get<Message[]>(`/tasks/${taskId}/messages`);
		} catch (e) {
			console.error('Failed to fetch messages:', e);
			messages = [];
		} finally {
			loading = false;
		}
	}

	async function sendMessage(taskId: string, content: string): Promise<Message | null> {
		try {
			if (useMockData) {
				const newMessage: Message = {
					id: 'msg-' + Date.now(),
					task_id: taskId,
					sender_type: 'user',
					sender_id: 'user-1',
					content,
					created_at: new Date().toISOString()
				};
				messages = [...messages, newMessage];

				// Simulate TARS response after a delay
				setTimeout(() => {
					const tarsReply: Message = {
						id: 'msg-' + (Date.now() + 1),
						task_id: taskId,
						sender_type: 'tars',
						content: 'Acknowledged. Processing your request.',
						created_at: new Date().toISOString()
					};
					messages = [...messages, tarsReply];
				}, 1000);

				return newMessage;
			}

			const { api } = await import('../api');
			const message = await api.post<Message>(`/tasks/${taskId}/messages`, { content });
			messages = [...messages, message];
			return message;
		} catch (e) {
			console.error('Failed to send message:', e);
			return null;
		}
	}

	function clear() {
		messages = [];
		currentTaskId = null;
	}

	return {
		get messages() { return messages; },
		get currentTaskId() { return currentTaskId; },
		get loading() { return loading; },
		fetchMessages,
		sendMessage,
		clear
	};
}

export const messagesStore = createMessagesStore();
