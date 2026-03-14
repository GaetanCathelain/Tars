import type { TimelineEvent } from '$shared/types/models';

interface EventsState {
	// repo_id → events (newest first)
	byRepo: Record<string, TimelineEvent[]>;
	loading: boolean;
}

const MAX_EVENTS_PER_REPO = 200;

function createEventsStore() {
	let state = $state<EventsState>({ byRepo: {}, loading: false });

	return {
		get byRepo() {
			return state.byRepo;
		},
		get loading() {
			return state.loading;
		},
		getEventsForRepo(repoId: string): TimelineEvent[] {
			return state.byRepo[repoId] ?? [];
		},
		setEvents(repoId: string, events: TimelineEvent[]) {
			state.byRepo = { ...state.byRepo, [repoId]: events };
		},
		addEvent(event: TimelineEvent) {
			const existing = state.byRepo[event.repo_id] ?? [];
			// Prepend and cap
			const updated = [event, ...existing].slice(0, MAX_EVENTS_PER_REPO);
			state.byRepo = { ...state.byRepo, [event.repo_id]: updated };
		},
		setLoading(loading: boolean) {
			state.loading = loading;
		}
	};
}

export const events = createEventsStore();
