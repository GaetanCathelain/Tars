import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import type { ListAgentsResponse, ListReposResponse, ListTasksResponse, SpawnAgentRequest } from '$shared/types/api';
import type { Agent, AgentPersona } from '$shared/types/models';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
	if (!locals.user) redirect(302, '/login');

	const sessionCookie = cookies.get('tars_session') ?? '';
	const selectedRepoId = url.searchParams.get('repoId');

	async function apiFetch<T>(path: string): Promise<T> {
		const res = await fetch(`${API_BASE}/api/v1${path}`, {
			headers: { Cookie: `tars_session=${sessionCookie}` }
		});
		if (!res.ok) return { repos: [], agents: [], tasks: [] } as T;
		return res.json() as Promise<T>;
	}

	const { repos } = await apiFetch<ListReposResponse>('/repos');
	const repoId = selectedRepoId ?? repos[0]?.id ?? null;

	let agentList: ListAgentsResponse = { agents: [] };
	let taskList: ListTasksResponse = { tasks: [] };

	if (repoId) {
		[agentList, taskList] = await Promise.all([
			apiFetch<ListAgentsResponse>(`/repos/${repoId}/agents`),
			apiFetch<ListTasksResponse>(`/repos/${repoId}/tasks`)
		]);
	}

	return { repos, agents: agentList.agents, tasks: taskList.tasks, selectedRepoId: repoId };
};

export const actions: Actions = {
	spawn: async ({ request, cookies, locals }) => {
		if (!locals.user) redirect(302, '/login');

		const sessionCookie = cookies.get('tars_session') ?? '';
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const name = (data.get('name') as string | null)?.trim();
		const persona = (data.get('persona') as AgentPersona | null) ?? 'general';
		const model = (data.get('model') as string | null) || 'claude-opus-4-6';
		const taskId = (data.get('task_id') as string | null) || undefined;
		const systemPrompt = (data.get('system_prompt') as string | null)?.trim() || undefined;

		if (!repoId || !name) return fail(400, { error: 'repoId and name are required.' });

		try {
			const body: SpawnAgentRequest = { name, persona, model };
			if (taskId) body.task_id = taskId;
			if (systemPrompt) body.system_prompt = systemPrompt;

			const res = await fetch(`${API_BASE}/api/v1/repos/${repoId}/agents`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Cookie: `tars_session=${sessionCookie}`
				},
				body: JSON.stringify(body)
			});

			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				const msg = (body as { error?: { message?: string } }).error?.message ?? `HTTP ${res.status}`;
				return fail(res.status, { error: msg });
			}

			const agent = await res.json() as Agent;
			return { success: true, agentId: agent.id };
		} catch (err) {
			return fail(500, { error: err instanceof Error ? err.message : 'Failed to spawn agent.' });
		}
	}
};
