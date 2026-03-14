import type { User } from '$shared/types/models';

interface AuthState {
	user: User | null;
	loading: boolean;
	error: string | null;
}

function createAuthStore() {
	let state = $state<AuthState>({
		user: null,
		loading: false,
		error: null
	});

	return {
		get user() {
			return state.user;
		},
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},
		get isAuthenticated() {
			return state.user !== null;
		},
		setUser(user: User | null) {
			state.user = user;
			state.error = null;
		},
		setLoading(loading: boolean) {
			state.loading = loading;
		},
		setError(error: string) {
			state.error = error;
			state.loading = false;
		},
		clear() {
			state.user = null;
			state.loading = false;
			state.error = null;
		}
	};
}

export const auth = createAuthStore();
