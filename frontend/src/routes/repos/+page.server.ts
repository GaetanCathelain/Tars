import { redirect, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { ListReposResponse, CreateRepoRequest, UpdateRepoRequest } from '$shared/types/api';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

async function apiRequest<T>(
	path: string,
	sessionCookie: string,
	init: RequestInit = {}
): Promise<T> {
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

export const load: PageServerLoad = async ({ locals, cookies }) => {
	if (!locals.user) redirect(302, '/login');

	const sessionCookie = cookies.get('tars_session') ?? '';
	let repoList: ListReposResponse = { repos: [] };
	try {
		repoList = await apiRequest<ListReposResponse>('/repos', sessionCookie);
	} catch {
		// Return empty list on error; client can retry
	}

	return { repos: repoList.repos };
};

export const actions: Actions = {
	create: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const name = (data.get('name') as string | null)?.trim();
		const github_url = (data.get('github_url') as string | null)?.trim();
		const path = (data.get('path') as string | null)?.trim();

		if (!name || !github_url || !path) {
			return fail(400, { error: 'All fields are required.' });
		}

		try {
			const body: CreateRepoRequest = { name, github_url, path };
			await apiRequest('/repos', sessionCookie, {
				method: 'POST',
				body: JSON.stringify(body)
			});
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to create repository.' });
		}

		return { success: true };
	},

	update: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const name = (data.get('name') as string | null)?.trim();
		const default_branch = (data.get('default_branch') as string | null)?.trim();

		if (!repoId) return fail(400, { error: 'Repository ID is required.' });

		const body: UpdateRepoRequest = {};
		if (name) body.name = name;
		if (default_branch) body.default_branch = default_branch;

		try {
			await apiRequest(`/repos/${repoId}`, sessionCookie, {
				method: 'PATCH',
				body: JSON.stringify(body)
			});
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to update repository.' });
		}

		return { success: true };
	},

	delete: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');
		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();
		const repoId = data.get('repoId') as string | null;

		if (!repoId) return fail(400, { error: 'Repository ID is required.' });

		try {
			await apiRequest(`/repos/${repoId}`, sessionCookie, { method: 'DELETE' });
		} catch (err) {
			return fail(422, { error: err instanceof Error ? err.message : 'Failed to delete repository.' });
		}

		return { success: true };
	}
};
