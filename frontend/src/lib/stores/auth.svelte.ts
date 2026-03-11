import type { User } from '$lib/types';
import { api } from '$lib/api';

const MOCK_MODE = false;

function decodeJwtPayload(token: string): { user_id: string; username: string } {
	const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
	return JSON.parse(atob(base64));
}

function createAuthStore() {
	let token = $state<string | null>(null);
	let user = $state<User | null>(null);

	function init() {
		if (typeof window === 'undefined') return;
		const savedToken = localStorage.getItem('auth_token');
		const savedUser = localStorage.getItem('auth_user');
		if (savedToken && savedUser) {
			token = savedToken;
			user = JSON.parse(savedUser);
		}
	}

	async function login(username: string, password: string): Promise<void> {
		if (MOCK_MODE) {
			const mockUser: User = {
				id: 'user-1',
				username,
				created_at: new Date().toISOString()
			};
			token = 'mock-jwt-token';
			user = mockUser;
			localStorage.setItem('auth_token', token);
			localStorage.setItem('auth_user', JSON.stringify(user));
			return;
		}
		const res = await api.post<{ token: string }>('/auth/login', { username, password });
		token = res.token;
		const payload = decodeJwtPayload(res.token);
		user = { id: payload.user_id, username: payload.username, created_at: new Date().toISOString() };
		localStorage.setItem('auth_token', res.token);
		localStorage.setItem('auth_user', JSON.stringify(user));
	}

	async function register(username: string, password: string): Promise<void> {
		if (MOCK_MODE) {
			return login(username, password);
		}
		const res = await api.post<{ token: string }>('/auth/register', { username, password });
		token = res.token;
		const payload = decodeJwtPayload(res.token);
		user = { id: payload.user_id, username: payload.username, created_at: new Date().toISOString() };
		localStorage.setItem('auth_token', res.token);
		localStorage.setItem('auth_user', JSON.stringify(user));
	}

	function logout() {
		token = null;
		user = null;
		localStorage.removeItem('auth_token');
		localStorage.removeItem('auth_user');
	}

	return {
		get token() { return token; },
		get user() { return user; },
		get isAuthenticated() { return !!token; },
		init,
		login,
		register,
		logout
	};
}

export const auth = createAuthStore();
