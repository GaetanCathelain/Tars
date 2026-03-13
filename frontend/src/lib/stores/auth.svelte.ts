import { goto } from '$app/navigation';
import { post } from '$lib/api';

export interface User {
	id: number;
	username: string;
	email: string;
	avatar_url: string;
}

class AuthStore {
	user = $state<User | null>(null);
	isAuthenticated = $derived(this.user !== null);

	setUser(user: User | null) {
		this.user = user;
	}

	async logout() {
		try {
			await post('/api/v1/auth/logout');
		} catch {
			// ignore errors on logout
		}
		this.user = null;
		goto('/login');
	}
}

export const auth = new AuthStore();
