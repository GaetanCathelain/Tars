import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

const PUBLIC_PATHS = ['/login', '/auth/callback'];

export const load: LayoutServerLoad = async ({ locals, url }) => {
	const isPublicPath = PUBLIC_PATHS.some((p) => url.pathname.startsWith(p));

	if (!locals.user && !isPublicPath) {
		redirect(302, '/login');
	}

	return {
		user: locals.user ?? null
	};
};
