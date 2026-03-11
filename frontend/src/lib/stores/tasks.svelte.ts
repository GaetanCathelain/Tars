import type { Task } from '../types';

const useMockData = true;

const mockTasks: Task[] = [
	{
		id: 'task-1',
		title: 'Set up authentication system',
		status: 'completed',
		created_by: 'user-1',
		created_at: '2026-03-10T14:00:00Z',
		updated_at: '2026-03-10T15:30:00Z'
	},
	{
		id: 'task-2',
		title: 'Build REST API endpoints',
		status: 'running',
		created_by: 'user-1',
		created_at: '2026-03-10T16:00:00Z',
		updated_at: '2026-03-11T00:15:00Z'
	},
	{
		id: 'task-3',
		title: 'Implement WebSocket streaming',
		status: 'open',
		created_by: 'user-1',
		created_at: '2026-03-11T00:30:00Z',
		updated_at: '2026-03-11T00:30:00Z'
	}
];

function createTasksStore() {
	let tasks = $state<Task[]>([]);
	let selectedTaskId = $state<string | null>(null);
	let loading = $state(false);

	async function fetchTasks() {
		loading = true;
		try {
			if (useMockData) {
				await new Promise((r) => setTimeout(r, 200));
				tasks = [...mockTasks];
				return;
			}

			const { api } = await import('../api');
			tasks = await api.get<Task[]>('/tasks');
		} catch (e) {
			console.error('Failed to fetch tasks:', e);
		} finally {
			loading = false;
		}
	}

	async function createTask(title: string): Promise<Task | null> {
		try {
			if (useMockData) {
				const newTask: Task = {
					id: 'task-' + Date.now(),
					title,
					status: 'open',
					created_by: 'user-1',
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};
				tasks = [...tasks, newTask];
				return newTask;
			}

			const { api } = await import('../api');
			const task = await api.post<Task>('/tasks', { title });
			tasks = [...tasks, task];
			return task;
		} catch (e) {
			console.error('Failed to create task:', e);
			return null;
		}
	}

	function selectTask(id: string) {
		selectedTaskId = id;
	}

	function updateTaskStatus(id: string, status: 'open' | 'running' | 'completed' | 'failed') {
		tasks = tasks.map((t) => (t.id === id ? { ...t, status, updated_at: new Date().toISOString() } : t));
	}

	return {
		get tasks() { return tasks; },
		get selectedTaskId() { return selectedTaskId; },
		get selectedTask() { return tasks.find((t) => t.id === selectedTaskId) ?? null; },
		get loading() { return loading; },
		fetchTasks,
		createTask,
		selectTask,
		updateTaskStatus
	};
}

export const tasksStore = createTasksStore();
