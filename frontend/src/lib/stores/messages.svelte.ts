import type { Message } from '$lib/types';
import { api } from '$lib/api';

function createMessagesStore() {
	let messages = $state<Message[]>([]);
	let currentTaskId = $state<string | null>(null);

	async function fetchMessages(taskId: string): Promise<void> {
		currentTaskId = taskId;
		messages = await api.get<Message[]>(`/tasks/${taskId}/messages`);
	}

	async function sendMessage(taskId: string, content: string): Promise<void> {
		const msg = await api.post<Message>(`/tasks/${taskId}/messages`, { content });
		messages = [...messages, msg];
	}

	function addMessage(message: Message): void {
		// Only add if it's for the currently viewed task
		if (message.task_id !== currentTaskId) return;
		// Avoid duplicates
		if (messages.some((m) => m.id === message.id)) return;
		messages = [...messages, message];
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
		addMessage,
		clear
	};
}

export const messagesStore = createMessagesStore();
