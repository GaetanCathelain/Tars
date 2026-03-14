import { redirect, error } from '@sveltejs/kit';
import type { Actions } from './$types';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

export const actions: Actions = {
	default: async ({ cookies, locals, params, url }) => {
		if (!locals.user) redirect(302, '/login');

		const sessionCookie = cookies.get('tars_session') ?? '';
		const { agentId } = params;
		const repoId = url.searchParams.get('repoId');
		if (!repoId) error(400, 'repoId required');

		await fetch(`${API_BASE}/api/v1/repos/${repoId}/agents/${agentId}/stop`, {
			method: 'POST',
			headers: { Cookie: `tars_session=${sessionCookie}` }
		});

		redirect(303, `/agents/${agentId}?repoId=${repoId}`);
	}
};
