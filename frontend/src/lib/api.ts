import { goto } from '$app/navigation';
import { PUBLIC_API_URL } from '$env/static/public';

const BASE_URL = PUBLIC_API_URL || '';

interface ApiResponse<T> {
	data: T;
}

interface ApiError {
	error: string;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const opts: RequestInit = {
		method,
		credentials: 'include',
		headers: {
			'Content-Type': 'application/json'
		}
	};

	if (body) {
		opts.body = JSON.stringify(body);
	}

	const res = await fetch(`${BASE_URL}${path}`, opts);

	if (res.status === 401) {
		goto('/login');
		throw new Error('Unauthorized');
	}

	if (res.status === 204) {
		return undefined as T;
	}

	if (!res.ok) {
		const err: ApiError = await res.json();
		throw new Error(err.error || `Request failed: ${res.status}`);
	}

	const json: ApiResponse<T> = await res.json();
	return json.data;
}

export function get<T>(path: string): Promise<T> {
	return request<T>('GET', path);
}

export function post<T>(path: string, body?: unknown): Promise<T> {
	return request<T>('POST', path, body);
}

export function patch<T>(path: string, body?: unknown): Promise<T> {
	return request<T>('PATCH', path, body);
}

export function del(path: string): Promise<void> {
	return request<void>('DELETE', path);
}

// Server-side fetch wrapper (uses SvelteKit fetch with cookie forwarding)
export function serverRequest<T>(
	fetchFn: typeof fetch,
	method: string,
	path: string,
	body?: unknown
): Promise<T> {
	const opts: RequestInit = {
		method,
		headers: {
			'Content-Type': 'application/json'
		}
	};

	if (body) {
		opts.body = JSON.stringify(body);
	}

	return fetchFn(`http://localhost:8080${path}`, opts).then(async (res) => {
		if (res.status === 204) return undefined as T;
		if (!res.ok) {
			const err: ApiError = await res.json().catch(() => ({ error: 'Request failed' }));
			throw new Error(err.error);
		}
		const json: ApiResponse<T> = await res.json();
		return json.data;
	});
}
