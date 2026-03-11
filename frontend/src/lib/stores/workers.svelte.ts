import type { WorkerSession } from '$lib/types';
import { api } from '$lib/api';

const MOCK_MODE = true;

const MOCK_WORKERS: Record<string, WorkerSession[]> = {
	'task-1': [
		{
			id: 'ws-auth-001',
			task_id: 'task-1',
			status: 'completed',
			command: 'claude-code --task "implement auth system"',
			exit_code: 0,
			started_at: '2025-03-10T14:01:00Z',
			finished_at: '2025-03-10T15:30:00Z'
		}
	],
	'task-2': [
		{
			id: 'ws-api-002',
			task_id: 'task-2',
			status: 'running',
			command: 'claude-code --task "build REST API endpoints"',
			started_at: '2025-03-10T16:01:00Z'
		}
	]
};

function createWorkersStore() {
	let sessions = $state<Record<string, WorkerSession[]>>({});

	async function fetchWorkers(taskId: string): Promise<void> {
		if (MOCK_MODE) {
			sessions = { ...sessions, [taskId]: MOCK_WORKERS[taskId] || [] };
			return;
		}
		const workers = await api.get<WorkerSession[]>(`/tasks/${taskId}/workers`);
		sessions = { ...sessions, [taskId]: workers };
	}

	async function spawnWorker(taskId: string, command: string): Promise<WorkerSession> {
		if (MOCK_MODE) {
			const worker: WorkerSession = {
				id: `ws-${Date.now()}`,
				task_id: taskId,
				status: 'running',
				command,
				started_at: new Date().toISOString()
			};
			sessions = { ...sessions, [taskId]: [...(sessions[taskId] || []), worker] };
			return worker;
		}
		const worker = await api.post<WorkerSession>(`/tasks/${taskId}/workers`, { command });
		sessions = { ...sessions, [taskId]: [...(sessions[taskId] || []), worker] };
		return worker;
	}

	async function killWorker(taskId: string, workerId: string): Promise<void> {
		if (MOCK_MODE) {
			sessions = {
				...sessions,
				[taskId]: (sessions[taskId] || []).map((w) =>
					w.id === workerId ? { ...w, status: 'failed' as const, finished_at: new Date().toISOString() } : w
				)
			};
			return;
		}
		await api.delete(`/tasks/${taskId}/workers/${workerId}`);
		await fetchWorkers(taskId);
	}

	function getWorkers(taskId: string): WorkerSession[] {
		return sessions[taskId] || [];
	}

	return {
		get sessions() { return sessions; },
		fetchWorkers,
		spawnWorker,
		killWorker,
		getWorkers
	};
}

export const workersStore = createWorkersStore();
