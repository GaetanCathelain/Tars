import type { AuthResponse, User } from '../types';

const useMockData = true;

function createAuthStore() {
	let token = $state<string | null>(null);
	let user = $state<User | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(false);

	// Load from localStorage on init
	if (typeof window !== 'undefined') {
		const savedToken = localStorage.getItem('tars_token');
		const savedUser = localStorage.getItem('tars_user');
		if (savedToken && savedUser) {
			token = savedToken;
			try {
				user = JSON.parse(savedUser);
			} catch {
				localStorage.removeItem('tars_token');
				localStorage.removeItem('tars_user');
			}
		}
	}

	function setAuth(authResponse: AuthResponse) {
		token = authResponse.token;
		user = authResponse.user;
		if (typeof window !== 'undefined') {
			localStorage.setItem('tars_token', authResponse.token);
			localStorage.setItem('tars_user', JSON.stringify(authResponse.user));
		}
	}

	async function login(username: string, password: string): Promise<boolean> {
		error = null;
		loading = true;
		try {
			if (useMockData) {
				// Simulate API delay
				await new Promise((r) => setTimeout(r, 300));
				if (username && password) {
					setAuth({
						token: 'mock-jwt-token-' + Date.now(),
						user: {
							id: 'user-1',
							username,
							created_at: new Date().toISOString()
						}
					});
					return true;
				}
				error = 'Invalid credentials';
				return false;
			}

			const { api } = await import('../api');
			const response = await api.post<AuthResponse>('/auth/login', { username, password });
			setAuth(response);
			return true;
		} catch (e: unknown) {
			error = (e as { error?: string })?.error || 'Login failed';
			return false;
		} finally {
			loading = false;
		}
	}

	async function register(username: string, password: string): Promise<boolean> {
		error = null;
		loading = true;
		try {
			if (useMockData) {
				await new Promise((r) => setTimeout(r, 300));
				if (username && password) {
					setAuth({
						token: 'mock-jwt-token-' + Date.now(),
						user: {
							id: 'user-' + Date.now(),
							username,
							created_at: new Date().toISOString()
						}
					});
					return true;
				}
				error = 'Registration failed';
				return false;
			}

			const { api } = await import('../api');
			const response = await api.post<AuthResponse>('/auth/register', { username, password });
			setAuth(response);
			return true;
		} catch (e: unknown) {
			error = (e as { error?: string })?.error || 'Registration failed';
			return false;
		} finally {
			loading = false;
		}
	}

	function logout() {
		token = null;
		user = null;
		if (typeof window !== 'undefined') {
			localStorage.removeItem('tars_token');
			localStorage.removeItem('tars_user');
		}
	}

	function clearError() {
		error = null;
	}

	return {
		get token() { return token; },
		get user() { return user; },
		get isAuthenticated() { return !!token && !!user; },
		get error() { return error; },
		get loading() { return loading; },
		login,
		register,
		logout,
		clearError
	};
}

export const authStore = createAuthStore();
