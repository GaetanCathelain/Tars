import type { Task } from '$lib/types';
import { api } from '$lib/api';

function createTasksStore() {
	let tasks = $state<Task[]>([]);
	let selectedTaskId = $state<string | null>(null);

	async function fetchTasks(): Promise<void> {
		tasks = await api.get<Task[]>('/tasks');
	}

	async function createTask(title: string): Promise<Task> {
		const task = await api.post<Task>('/tasks', { title });
		tasks = [...tasks, task];
		return task;
	}

	function selectTask(id: string | null) {
		selectedTaskId = id;
	}

	function updateTaskStatus(taskId: string, status: string): void {
		tasks = tasks.map((t) =>
			t.id === taskId
				? { ...t, status: status as Task['status'], updated_at: new Date().toISOString() }
				: t
		);
	}

	return {
		get tasks() { return tasks; },
		get selectedTaskId() { return selectedTaskId; },
		get selectedTask() { return tasks.find((t) => t.id === selectedTaskId) ?? null; },
		fetchTasks,
		createTask,
		selectTask,
		updateTaskStatus
	};
}

export const tasksStore = createTasksStore();
