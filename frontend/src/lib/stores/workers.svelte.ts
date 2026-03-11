import type { WorkerSession } from '$lib/types';
import { api } from '$lib/api';

function createWorkersStore() {
	let sessions = $state<Record<string, WorkerSession[]>>({});

	async function fetchWorkers(taskId: string): Promise<void> {
		const workers = await api.get<WorkerSession[]>(`/tasks/${taskId}/workers`);
		sessions = { ...sessions, [taskId]: workers };
	}

	async function spawnWorker(taskId: string, command: string): Promise<WorkerSession> {
		const worker = await api.post<WorkerSession>(`/tasks/${taskId}/workers`, { command });
		sessions = { ...sessions, [taskId]: [...(sessions[taskId] || []), worker] };
		return worker;
	}

	async function killWorker(taskId: string, workerId: string): Promise<void> {
		await api.delete(`/tasks/${taskId}/workers/${workerId}`);
		await fetchWorkers(taskId);
	}

	function getWorkers(taskId: string): WorkerSession[] {
		return sessions[taskId] || [];
	}

	function addWorker(worker: WorkerSession): void {
		const taskId = worker.task_id;
		const existing = sessions[taskId] || [];
		// Avoid duplicates
		if (existing.some((w) => w.id === worker.id)) return;
		sessions = { ...sessions, [taskId]: [...existing, worker] };
	}

	function updateWorkerStatus(sessionId: string, status: string, exitCode?: number): void {
		const updated: Record<string, WorkerSession[]> = {};
		for (const [taskId, workers] of Object.entries(sessions)) {
			updated[taskId] = workers.map((w) => {
				if (w.id === sessionId) {
					return {
						...w,
						status: status as WorkerSession['status'],
						exit_code: exitCode ?? w.exit_code,
						finished_at: status !== 'running' ? new Date().toISOString() : w.finished_at
					};
				}
				return w;
			});
		}
		sessions = updated;
	}

	return {
		get sessions() { return sessions; },
		fetchWorkers,
		spawnWorker,
		killWorker,
		getWorkers,
		addWorker,
		updateWorkerStatus
	};
}

export const workersStore = createWorkersStore();
