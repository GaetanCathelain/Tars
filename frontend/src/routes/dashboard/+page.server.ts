import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ListReposResponse, ListAgentsResponse, ListTasksResponse } from '$shared/types/api';
import type { Agent, Task, Repo } from '$shared/types/models';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

export const load: PageServerLoad = async ({ locals, cookies }) => {
	if (!locals.user) redirect(302, '/login');

	const sessionCookie = cookies.get('tars_session') ?? '';

	async function apiFetch<T>(path: string): Promise<T | null> {
		try {
			const res = await fetch(`${API_BASE}/api/v1${path}`, {
				headers: { Cookie: `tars_session=${sessionCookie}` }
			});
			if (!res.ok) return null;
			return res.json() as Promise<T>;
		} catch {
			return null;
		}
	}

	const reposData = await apiFetch<ListReposResponse>('/repos');
	const repos: Repo[] = reposData?.repos ?? [];

	// Fetch agents + tasks for all repos (up to first 3 to avoid hammering)
	const targetRepos = repos.slice(0, 3);

	const [agentResults, taskResults] = await Promise.all([
		Promise.all(
			targetRepos.map((r) =>
				apiFetch<ListAgentsResponse>(`/repos/${r.id}/agents`).then((d) => d?.agents ?? [])
			)
		),
		Promise.all(
			targetRepos.map((r) =>
				apiFetch<ListTasksResponse>(`/repos/${r.id}/tasks`).then((d) => d?.tasks ?? [])
			)
		)
	]);

	const agents: Agent[] = agentResults.flat();
	const tasks: Task[] = taskResults.flat();

	return { repos, agents, tasks };
};
