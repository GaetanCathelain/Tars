import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';
import type { User } from '$shared/types/models';

const API_BASE = process.env['API_URL'] ?? 'http://localhost:8090';

// Routes that don't require authentication
const PUBLIC_ROUTES = ['/login', '/auth/callback'];

function isPublicRoute(pathname: string): boolean {
	return PUBLIC_ROUTES.some((r) => pathname === r || pathname.startsWith(r + '/'));
}

export const handle: Handle = async ({ event, resolve }) => {
	const sessionCookie = event.cookies.get('tars_session');

	if (sessionCookie) {
		try {
			const res = await fetch(`${API_BASE}/api/v1/auth/me`, {
				headers: { Cookie: `tars_session=${sessionCookie}` }
			});
			if (res.ok) {
				event.locals.user = (await res.json()) as User;
			} else if (res.status === 401) {
				// Session expired — clear the cookie
				event.cookies.delete('tars_session', { path: '/' });
			}
		} catch {
			// Network error — leave user as undefined, don't clear cookie
		}
	}

	const pathname = event.url.pathname;

	// Redirect unauthenticated users away from protected routes
	if (!event.locals.user && !isPublicRoute(pathname) && pathname !== '/') {
		redirect(302, `/login`);
	}

	// Redirect authenticated users away from login page
	if (event.locals.user && pathname === '/login') {
		redirect(302, '/dashboard');
	}

	return resolve(event);
};
