import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const sessionCookie = event.cookies.get('tars_session');

	if (sessionCookie) {
		try {
			const res = await fetch('http://localhost:8080/api/v1/auth/me', {
				headers: {
					Cookie: `tars_session=${sessionCookie}`
				}
			});

			if (res.ok) {
				const json = await res.json();
				event.locals.user = json.data;
			}
		} catch {
			// Backend unreachable — treat as unauthenticated
		}
	}

	return resolve(event);
};
