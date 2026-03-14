import type { Task } from '$shared/types/models';

interface TasksState {
	tasks: Task[];
	loading: boolean;
	error: string | null;
}

function createTasksStore() {
	let state = $state<TasksState>({
		tasks: [],
		loading: false,
		error: null
	});

	return {
		get tasks() {
			return state.tasks;
		},
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},
		setTasks(tasks: Task[]) {
			state.tasks = tasks;
			state.error = null;
		},
		addTask(task: Task) {
			if (!state.tasks.find((t) => t.id === task.id)) {
				state.tasks = [...state.tasks, task];
			}
		},
		updateTask(updated: Task) {
			state.tasks = state.tasks.map((t) => (t.id === updated.id ? updated : t));
		},
		removeTask(id: string) {
			state.tasks = state.tasks.filter((t) => t.id !== id);
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

export const tasks = createTasksStore();
