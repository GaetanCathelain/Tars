import type { ApiErrorResponse } from '$shared/types/api';

// PUBLIC_API_URL is injected at build time via Vite's env handling.
// In SvelteKit, we access it via $env/static/public in .svelte files.
// For use in .ts lib files that may run on both client and server,
// we read from the Vite import.meta.env object.
function getApiBase(): string {
	// During SSR, fall back to localhost
	if (typeof import.meta !== 'undefined' && import.meta.env) {
		return (import.meta.env['PUBLIC_API_URL'] as string) ?? 'http://localhost:8090';
	}
	return 'http://localhost:8090';
}

export class ApiError extends Error {
	constructor(
		public readonly code: string,
		message: string,
		public readonly status: number,
		public readonly details?: Record<string, unknown>
	) {
		super(message);
		this.name = 'ApiError';
	}
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const url = `${getApiBase()}/api/v1${path}`;
	const res = await fetch(url, {
		...init,
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json',
			...init.headers
		}
	});

	if (!res.ok) {
		let errorBody: ApiErrorResponse | undefined;
		try {
			errorBody = (await res.json()) as ApiErrorResponse;
		} catch {
			// non-JSON error body
		}
		throw new ApiError(
			errorBody?.error.code ?? 'UNKNOWN',
			errorBody?.error.message ?? `HTTP ${res.status}`,
			res.status,
			errorBody?.error.details
		);
	}

	if (res.status === 204) {
		return undefined as T;
	}

	return res.json() as Promise<T>;
}

export const api = {
	get<T>(path: string): Promise<T> {
		return request<T>(path);
	},
	post<T>(path: string, body?: unknown): Promise<T> {
		return request<T>(path, {
			method: 'POST',
			body: body !== undefined ? JSON.stringify(body) : undefined
		});
	},
	patch<T>(path: string, body: unknown): Promise<T> {
		return request<T>(path, { method: 'PATCH', body: JSON.stringify(body) });
	},
	delete<T = void>(path: string): Promise<T> {
		return request<T>(path, { method: 'DELETE' });
	}
};
