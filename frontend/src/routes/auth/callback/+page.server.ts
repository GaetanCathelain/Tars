import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

// The backend handles the OAuth callback at /api/v1/auth/github/callback and sets
// the tars_session cookie, then redirects here. By the time we land on this page
// the cookie is already set. We just forward to the dashboard.
export const load: PageServerLoad = async () => {
	redirect(302, '/dashboard');
};
