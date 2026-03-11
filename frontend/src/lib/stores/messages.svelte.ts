import type { Message } from '$lib/types';
import { api } from '$lib/api';

const MOCK_MODE = false;

const MOCK_MESSAGES: Record<string, Message[]> = {
	'task-1': [
		{
			id: 'msg-1',
			task_id: 'task-1',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Set up JWT authentication with login and register endpoints.',
			created_at: '2025-03-10T14:00:00Z'
		},
		{
			id: 'msg-2',
			task_id: 'task-1',
			sender_type: 'tars',
			content: 'Understood. I\'ll create the auth middleware, login/register handlers, and JWT token generation. Spawning a worker now.',
			created_at: '2025-03-10T14:00:30Z'
		},
		{
			id: 'msg-3',
			task_id: 'task-1',
			sender_type: 'system',
			content: 'Worker spawned: claude-code (session ws-auth-001)',
			created_at: '2025-03-10T14:01:00Z'
		},
		{
			id: 'msg-4',
			task_id: 'task-1',
			sender_type: 'system',
			content: 'Worker completed successfully. Branch: tars/auth-system. PR #12 created.',
			created_at: '2025-03-10T15:30:00Z'
		},
		{
			id: 'msg-5',
			task_id: 'task-1',
			sender_type: 'tars',
			content: 'Auth system is done. Created JWT middleware, bcrypt password hashing, login/register endpoints, and tests. PR #12 is ready for review.',
			created_at: '2025-03-10T15:30:30Z'
		}
	],
	'task-2': [
		{
			id: 'msg-6',
			task_id: 'task-2',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Build CRUD endpoints for tasks and messages. Include pagination.',
			created_at: '2025-03-10T16:00:00Z'
		},
		{
			id: 'msg-7',
			task_id: 'task-2',
			sender_type: 'tars',
			content: 'On it. I\'ll set up the task and message handlers with proper pagination, validation, and error handling.',
			created_at: '2025-03-10T16:00:15Z'
		},
		{
			id: 'msg-8',
			task_id: 'task-2',
			sender_type: 'system',
			content: 'Worker spawned: claude-code (session ws-api-002)',
			created_at: '2025-03-10T16:01:00Z'
		},
		{
			id: 'msg-9',
			task_id: 'task-2',
			sender_type: 'tars',
			content: 'Worker is still running. Task endpoints are done, now working on message endpoints and pagination logic.',
			created_at: '2025-03-10T16:45:00Z'
		}
	],
	'task-3': [
		{
			id: 'msg-10',
			task_id: 'task-3',
			sender_type: 'user',
			sender_id: 'user-1',
			content: 'Add WebSocket support for real-time task updates and message streaming.',
			created_at: '2025-03-10T17:00:00Z'
		}
	]
};

function createMessagesStore() {
	let messages = $state<Message[]>([]);
	let currentTaskId = $state<string | null>(null);

	async function fetchMessages(taskId: string): Promise<void> {
		currentTaskId = taskId;
		if (MOCK_MODE) {
			messages = MOCK_MESSAGES[taskId] || [];
			return;
		}
		messages = await api.get<Message[]>(`/tasks/${taskId}/messages`);
	}

	async function sendMessage(taskId: string, content: string): Promise<void> {
		if (MOCK_MODE) {
			const newMsg: Message = {
				id: `msg-${Date.now()}`,
				task_id: taskId,
				sender_type: 'user',
				sender_id: 'user-1',
				content,
				created_at: new Date().toISOString()
			};
			messages = [...messages, newMsg];

			// Mock TARS response after a short delay
			setTimeout(() => {
				const tarsReply: Message = {
					id: `msg-${Date.now() + 1}`,
					task_id: taskId,
					sender_type: 'tars',
					content: 'Got it. Let me look into that and get back to you.',
					created_at: new Date().toISOString()
				};
				messages = [...messages, tarsReply];
			}, 1000);
			return;
		}
		const msg = await api.post<Message>(`/tasks/${taskId}/messages`, { content });
		messages = [...messages, msg];
	}

	function clear() {
		messages = [];
		currentTaskId = null;
	}

	return {
		get messages() { return messages; },
		get currentTaskId() { return currentTaskId; },
		fetchMessages,
		sendMessage,
		clear
	};
}

export const messagesStore = createMessagesStore();
