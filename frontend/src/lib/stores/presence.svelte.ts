import type { PresenceUser } from '$shared/types/models';

interface PresenceState {
	// repo_id → users
	byRepo: Record<string, PresenceUser[]>;
}

function createPresenceStore() {
	let state = $state<PresenceState>({ byRepo: {} });

	return {
		get byRepo() {
			return state.byRepo;
		},
		getUsersForRepo(repoId: string): PresenceUser[] {
			return state.byRepo[repoId] ?? [];
		},
		setSnapshot(repoId: string, users: PresenceUser[]) {
			state.byRepo = { ...state.byRepo, [repoId]: users };
		},
		clear() {
			state.byRepo = {};
		}
	};
}

export const presence = createPresenceStore();
