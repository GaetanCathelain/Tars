import { redirect, error, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import type { Agent, AgentLogsResponse, RepoDiff, PresenceState, MergeResult } from '$shared/types/models';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

export const load: PageServerLoad = async ({ locals, cookies, params, url }) => {
	if (!locals.user) redirect(302, '/login');

	const sessionCookie = cookies.get('tars_session') ?? '';
	const { agentId } = params;

	// repoId is required — passed as a query param from the listing
	const repoId = url.searchParams.get('repoId');
	if (!repoId) error(400, 'repoId query param required');

	async function apiFetch<T>(path: string): Promise<T> {
		const res = await fetch(`${API_BASE}/api/v1${path}`, {
			headers: { Cookie: `tars_session=${sessionCookie}` }
		});
		if (!res.ok) error(res.status, `API error: ${res.status}`);
		return res.json() as Promise<T>;
	}

	const [agent, logs, presence, diff] = await Promise.all([
		apiFetch<Agent>(`/repos/${repoId}/agents/${agentId}`),
		apiFetch<AgentLogsResponse>(`/repos/${repoId}/agents/${agentId}/logs?limit=500`),
		apiFetch<PresenceState>(`/repos/${repoId}/presence`).catch(() => ({ repo_id: repoId, users: [] } as PresenceState)),
		apiFetch<RepoDiff>(`/repos/${repoId}/agents/${agentId}/diff`).catch(() => null)
	]);

	return { agent, repoId, initialLines: logs.lines, presence, diff };
};

export const actions: Actions = {
	merge: async ({ request, cookies, locals, params }) => {
		if (!locals.user) redirect(302, '/login');

		const sessionCookie = cookies.get('tars_session') ?? '';
		const { agentId } = params;
		const data = await request.formData();

		const repoId = data.get('repoId') as string | null;
		const targetBranch = (data.get('targetBranch') as string | null)?.trim() || 'main';
		const strategy = (data.get('strategy') as string | null) || 'squash';
		const commitMessage = (data.get('commitMessage') as string | null)?.trim() || undefined;

		if (!repoId) return fail(400, { error: 'repoId is required.' });

		try {
			const body: Record<string, string> = { target_branch: targetBranch, strategy };
			if (commitMessage) body['commit_message'] = commitMessage;

			const res = await fetch(`${API_BASE}/api/v1/repos/${repoId}/agents/${agentId}/merge`, {
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
				if (res.status === 409) return fail(409, { error: `Merge conflict: ${msg}` });
				return fail(res.status, { error: msg });
			}

			const result = await res.json() as MergeResult;
			return { success: true, mergeResult: result };
		} catch (err) {
			return fail(500, { error: err instanceof Error ? err.message : 'Merge failed.' });
		}
	}
};
