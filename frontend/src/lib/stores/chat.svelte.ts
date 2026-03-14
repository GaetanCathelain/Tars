export interface ChatMessage {
	id: string;
	agentId: string;
	role: 'user';
	text: string;
	sentAt: string; // ISO 8601
}

interface ChatState {
	// agentId → messages sent by the user
	byAgent: Record<string, ChatMessage[]>;
	sending: boolean;
	error: string | null;
}

function createChatStore() {
	let state = $state<ChatState>({ byAgent: {}, sending: false, error: null });

	return {
		get byAgent() {
			return state.byAgent;
		},
		get sending() {
			return state.sending;
		},
		get error() {
			return state.error;
		},
		getMessages(agentId: string): ChatMessage[] {
			return state.byAgent[agentId] ?? [];
		},
		addMessage(msg: ChatMessage) {
			const existing = state.byAgent[msg.agentId] ?? [];
			state.byAgent = { ...state.byAgent, [msg.agentId]: [...existing, msg] };
		},
		setSending(sending: boolean) {
			state.sending = sending;
		},
		setError(error: string | null) {
			state.error = error;
		},
		clearError() {
			state.error = null;
		}
	};
}

export const chat = createChatStore();
