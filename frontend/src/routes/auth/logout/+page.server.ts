import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

export const actions: Actions = {
	default: async ({ cookies, fetch }) => {
		const sessionCookie = cookies.get('tars_session');

		if (sessionCookie) {
			try {
				await fetch(`${API_BASE}/api/v1/auth/logout`, {
					method: 'POST',
					headers: { Cookie: `tars_session=${sessionCookie}` }
				});
			} catch {
				// Best-effort — clear cookie regardless
			}
			cookies.delete('tars_session', { path: '/' });
		}

		redirect(302, '/login');
	}
};
