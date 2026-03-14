import type { Agent, AgentStatus, AgentLogLine } from '$shared/types/models';

interface AgentWithOutput extends Agent {
	outputLines: AgentLogLine[];
}

interface AgentsState {
	agents: AgentWithOutput[];
	loading: boolean;
	error: string | null;
}

function createAgentsStore() {
	let state = $state<AgentsState>({
		agents: [],
		loading: false,
		error: null
	});

	return {
		get agents() {
			return state.agents;
		},
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},
		setAgents(agents: Agent[]) {
			state.agents = agents.map((a) => ({ ...a, outputLines: [] }));
			state.error = null;
		},
		addAgent(agent: Agent) {
			if (!state.agents.find((a) => a.id === agent.id)) {
				state.agents = [...state.agents, { ...agent, outputLines: [] }];
			}
		},
		updateStatus(agentId: string, status: AgentStatus, exitCode: number | null) {
			state.agents = state.agents.map((a) =>
				a.id === agentId
					? { ...a, status, stopped_at: status === 'stopped' || status === 'crashed' ? new Date().toISOString() : a.stopped_at }
					: a
			);
			// exitCode is available if needed by consumers via the WS message directly
			void exitCode;
		},
		appendOutput(agentId: string, line: AgentLogLine) {
			state.agents = state.agents.map((a) => {
				if (a.id !== agentId) return a;
				// Keep last 5000 lines in memory
				const lines = a.outputLines.length >= 5000
					? [...a.outputLines.slice(-4999), line]
					: [...a.outputLines, line];
				return { ...a, outputLines: lines };
			});
		},
		removeAgent(id: string) {
			state.agents = state.agents.filter((a) => a.id !== id);
		},
		setLoading(loading: boolean) {
			state.loading = loading;
		},
		setError(error: string) {
			state.error = error;
			state.loading = false;
		}
	};
}

export const agents = createAgentsStore();
