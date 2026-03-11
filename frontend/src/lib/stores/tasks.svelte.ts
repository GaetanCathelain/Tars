import type { Task } from '$lib/types';
import { api } from '$lib/api';

const MOCK_MODE = false;

const MOCK_TASKS: Task[] = [
	{
		id: 'task-1',
		title: 'Set up authentication system',
		status: 'completed',
		created_by: 'user-1',
		created_at: '2025-03-10T14:00:00Z',
		updated_at: '2025-03-10T15:30:00Z'
	},
	{
		id: 'task-2',
		title: 'Build REST API endpoints',
		status: 'running',
		created_by: 'user-1',
		created_at: '2025-03-10T16:00:00Z',
		updated_at: '2025-03-10T16:45:00Z'
	},
	{
		id: 'task-3',
		title: 'WebSocket real-time updates',
		status: 'open',
		created_by: 'user-1',
		created_at: '2025-03-10T17:00:00Z',
		updated_at: '2025-03-10T17:00:00Z'
	}
];

function createTasksStore() {
	let tasks = $state<Task[]>([]);
	let selectedTaskId = $state<string | null>(null);

	async function fetchTasks(): Promise<void> {
		if (MOCK_MODE) {
			tasks = [...MOCK_TASKS];
			return;
		}
		tasks = await api.get<Task[]>('/tasks');
	}

	async function createTask(title: string): Promise<Task> {
		if (MOCK_MODE) {
			const newTask: Task = {
				id: `task-${Date.now()}`,
				title,
				status: 'open',
				created_by: 'user-1',
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
			tasks = [...tasks, newTask];
			return newTask;
		}
		const task = await api.post<Task>('/tasks', { title });
		tasks = [...tasks, task];
		return task;
	}

	function selectTask(id: string | null) {
		selectedTaskId = id;
	}

	return {
		get tasks() { return tasks; },
		get selectedTaskId() { return selectedTaskId; },
		get selectedTask() { return tasks.find((t) => t.id === selectedTaskId) ?? null; },
		fetchTasks,
		createTask,
		selectTask
	};
}

export const tasksStore = createTasksStore();
