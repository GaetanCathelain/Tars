const BASE_URL = '/api';

class ApiClient {
	private getToken(): string | null {
		if (typeof window === 'undefined') return null;
		return localStorage.getItem('auth_token');
	}

	private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
		const token = this.getToken();
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...((options.headers as Record<string, string>) || {})
		};
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		const response = await fetch(`${BASE_URL}${path}`, {
			...options,
			headers
		});

		if (response.status === 401) {
			localStorage.removeItem('auth_token');
			localStorage.removeItem('auth_user');
			window.location.href = '/login';
			throw new Error('Unauthorized');
		}

		if (!response.ok) {
			const error = await response.json().catch(() => ({ message: 'Request failed' }));
			throw new Error(error.message || `HTTP ${response.status}`);
		}

		return response.json();
	}

	get<T>(path: string): Promise<T> {
		return this.request<T>(path);
	}

	post<T>(path: string, body?: unknown): Promise<T> {
		return this.request<T>(path, {
			method: 'POST',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	put<T>(path: string, body?: unknown): Promise<T> {
		return this.request<T>(path, {
			method: 'PUT',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	delete<T>(path: string): Promise<T> {
		return this.request<T>(path, { method: 'DELETE' });
	}
}

export const api = new ApiClient();
