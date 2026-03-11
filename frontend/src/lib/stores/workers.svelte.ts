import type { WorkerSession } from '../types';

interface WorkerOutput {
	id: string;
	data: string; // base64
	created_at: string;
}

function createWorkersStore() {
	let workers = $state<Map<string, WorkerSession[]>>(new Map());

	function getWorkersForTask(taskId: string): WorkerSession[] {
		return workers.get(taskId) ?? [];
	}

	function getWorker(sessionId: string): WorkerSession | undefined {
		for (const sessions of workers.values()) {
			const found = sessions.find((w) => w.id === sessionId);
			if (found) return found;
		}
		return undefined;
	}

	function onWorkerStart(session: WorkerSession) {
		const taskWorkers = workers.get(session.task_id) ?? [];
		// Avoid duplicates
		if (taskWorkers.find((w) => w.id === session.id)) return;
		const updated = new Map(workers);
		updated.set(session.task_id, [...taskWorkers, session]);
		workers = updated;
	}

	function onWorkerEnd(sessionId: string, exitCode: number) {
		const updated = new Map(workers);
		for (const [taskId, sessions] of updated) {
			const idx = sessions.findIndex((w) => w.id === sessionId);
			if (idx !== -1) {
				const updatedSessions = [...sessions];
				updatedSessions[idx] = {
					...updatedSessions[idx],
					status: exitCode === 0 ? 'completed' : 'failed',
					exit_code: exitCode,
					finished_at: new Date().toISOString()
				};
				updated.set(taskId, updatedSessions);
				break;
			}
		}
		workers = updated;
	}

	async function spawnWorker(taskId: string, prompt: string): Promise<WorkerSession | null> {
		try {
			const { api } = await import('../api');
			const session = await api.post<WorkerSession>(`/tasks/${taskId}/workers`, { prompt });
			onWorkerStart(session);
			return session;
		} catch (e) {
			console.error('Failed to spawn worker:', e);
			return null;
		}
	}

	async function killWorker(sessionId: string): Promise<boolean> {
		try {
			const { api } = await import('../api');
			await api.del(`/workers/${sessionId}`);
			onWorkerEnd(sessionId, -1);
			return true;
		} catch (e) {
			console.error('Failed to kill worker:', e);
			return false;
		}
	}

	async function fetchOutput(sessionId: string): Promise<Uint8Array[]> {
		try {
			const { api } = await import('../api');
			const outputs = await api.get<WorkerOutput[]>(`/workers/${sessionId}/output`);
			return outputs.map((o) => Uint8Array.from(atob(o.data), (c) => c.charCodeAt(0)));
		} catch (e) {
			console.error('Failed to fetch worker output:', e);
			return [];
		}
	}

	function clear() {
		workers = new Map();
	}

	return {
		get workers() { return workers; },
		getWorkersForTask,
		getWorker,
		onWorkerStart,
		onWorkerEnd,
		spawnWorker,
		killWorker,
		fetchOutput,
		clear
	};
}

export const workersStore = createWorkersStore();
