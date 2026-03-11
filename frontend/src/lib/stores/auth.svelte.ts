import type { User, AuthResponse } from '$lib/types';
import { api } from '$lib/api';

const MOCK_MODE = true;

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
		const res = await api.post<AuthResponse>('/auth/login', { username, password });
		token = res.token;
		user = res.user;
		localStorage.setItem('auth_token', res.token);
		localStorage.setItem('auth_user', JSON.stringify(res.user));
	}

	async function register(username: string, password: string): Promise<void> {
		if (MOCK_MODE) {
			return login(username, password);
		}
		const res = await api.post<AuthResponse>('/auth/register', { username, password });
		token = res.token;
		user = res.user;
		localStorage.setItem('auth_token', res.token);
		localStorage.setItem('auth_user', JSON.stringify(res.user));
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
