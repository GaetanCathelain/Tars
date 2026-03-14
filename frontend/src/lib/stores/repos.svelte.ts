import type { Repo } from '$shared/types/models';

interface ReposState {
	repos: Repo[];
	loading: boolean;
	error: string | null;
}

function createReposStore() {
	let state = $state<ReposState>({
		repos: [],
		loading: false,
		error: null
	});

	return {
		get repos() {
			return state.repos;
		},
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},
		setRepos(repos: Repo[]) {
			state.repos = repos;
			state.error = null;
		},
		addRepo(repo: Repo) {
			state.repos = [...state.repos, repo];
		},
		updateRepo(updated: Repo) {
			state.repos = state.repos.map((r) => (r.id === updated.id ? updated : r));
		},
		removeRepo(id: string) {
			state.repos = state.repos.filter((r) => r.id !== id);
		},
		setLoading(loading: boolean) {
			state.loading = loading;
		},
		setError(error: string) {
			state.error = error;
			state.loading = false;
		}
	};
}

export const repos = createReposStore();
