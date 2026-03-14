import { redirect, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { ListTasksResponse, CreateTaskRequest, UpdateTaskRequest, ListReposResponse } from '$shared/types/api';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

async function apiFetch<T>(path: string, sessionCookie: string, init: RequestInit = {}): Promise<T> {
	const res = await fetch(`${API_BASE}/api/v1${path}`, {
		...init,
		headers: {
			'Content-Type': 'application/json',
			Cookie: `tars_session=${sessionCookie}`,
			...init.headers
		}
	});
	if (!res.ok) {
		const body = await res.json().catch(() => ({}));
		throw new Error((body as { error?: { message?: string } }).error?.message ?? `HTTP ${res.status}`);
	}
	if (res.status === 204) return undefined as T;
	return res.json() as Promise<T>;
}

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
	if (!locals.user) redirect(302, '/login');

	const sessionCookie = cookies.get('tars_session') ?? '';
	const selectedRepoId = url.searchParams.get('repoId');

	const { repos } = await apiFetch<ListReposResponse>('/repos', sessionCookie).catch(() => ({ repos: [] }));
	const repoId = selectedRepoId ?? repos[0]?.id ?? null;

	let tasks: ListTasksResponse = { tasks: [] };
	if (repoId) {
		tasks = await apiFetch<ListTasksResponse>(`/repos/${repoId}/tasks`, sessionCookie).catch(() => ({ tasks: [] }));
	}

	return { repos, tasks: tasks.tasks, selectedRepoId: repoId };
};

export const actions: Actions = {
	create: async ({ request, cookies, locals, url }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const title = (data.get('title') as string | null)?.trim();
		const description = (data.get('description') as string | null)?.trim() ?? '';
		const priority = parseInt(data.get('priority') as string ?? '3', 10);

		if (!repoId || !title) return fail(400, { error: 'repoId and title are required.' });

		try {
			const body: CreateTaskRequest = { title, description, priority };
			await apiFetch(`/repos/${repoId}/tasks`, sessionCookie, {
				method: 'POST',
				body: JSON.stringify(body)
			});
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to create task.' });
		}
		return { success: true };
	},

	move: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const taskId = data.get('taskId') as string | null;
		const status = data.get('status') as string | null;

		if (!repoId || !taskId || !status) return fail(400, { error: 'Missing fields.' });

		try {
			const body: UpdateTaskRequest = { status: status as UpdateTaskRequest['status'] };
			await apiFetch(`/repos/${repoId}/tasks/${taskId}`, sessionCookie, {
				method: 'PATCH',
				body: JSON.stringify(body)
			});
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to move task.' });
		}
		return { success: true };
	},

	update: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const taskId = data.get('taskId') as string | null;
		const title = (data.get('title') as string | null)?.trim();
		const description = (data.get('description') as string | null)?.trim();
		const priority = data.get('priority') ? parseInt(data.get('priority') as string, 10) : undefined;

		if (!repoId || !taskId) return fail(400, { error: 'Missing fields.' });

		try {
			const body: UpdateTaskRequest = {};
			if (title) body.title = title;
			if (description !== undefined) body.description = description;
			if (priority) body.priority = priority;
			await apiFetch(`/repos/${repoId}/tasks/${taskId}`, sessionCookie, {
				method: 'PATCH',
				body: JSON.stringify(body)
			});
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to update task.' });
		}
		return { success: true };
	},

	delete: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const taskId = data.get('taskId') as string | null;
		if (!repoId || !taskId) return fail(400, { error: 'Missing fields.' });

		try {
			await apiFetch(`/repos/${repoId}/tasks/${taskId}`, sessionCookie, { method: 'DELETE' });
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to delete task.' });
		}
		return { success: true };
	}
};
