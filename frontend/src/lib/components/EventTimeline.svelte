<script lang="ts">
	import { onMount } from 'svelte';
	import { events } from '$lib/stores/events.svelte';
	import { api, ApiError } from '$lib/utils/api';
	import type { TimelineEvent } from '$shared/types/models';
	import { Bot, User, Zap, GitMerge, CheckSquare, Trash2, Plus, AlertCircle } from 'lucide-svelte';

	interface Props {
		repoId: string;
		class?: string;
	}

	let { repoId, class: className = '' }: Props = $props();

	let error = $state<string | null>(null);

	onMount(async () => {
		if (!repoId) return;
		events.setLoading(true);
		try {
			const res = await api.get<{ events: TimelineEvent[]; has_more: boolean }>(
				`/repos/${repoId}/events?limit=50`
			);
			events.setEvents(repoId, res.events);
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to load events.';
		} finally {
			events.setLoading(false);
		}
	});

	const repoEvents: TimelineEvent[] = $derived(events.getEventsForRepo(repoId));

	function formatTime(iso: string): string {
		try {
			const d = new Date(iso);
			return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} catch {
			return '';
		}
	}

	function formatDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString([], { month: 'short', day: 'numeric' });
		} catch {
			return '';
		}
	}

	interface EventMeta {
		icon: typeof Bot;
		color: string;
		label: string;
	}

	function eventMeta(type: string): EventMeta {
		const map: Record<string, EventMeta> = {
			'agent.spawned':  { icon: Bot,       color: 'text-green-400',  label: 'Agent spawned' },
			'agent.stopped':  { icon: Bot,       color: 'text-zinc-500',   label: 'Agent stopped' },
			'agent.crashed':  { icon: AlertCircle, color: 'text-red-400',  label: 'Agent crashed' },
			'agent.merged':   { icon: GitMerge,  color: 'text-purple-400', label: 'Agent merged' },
			'task.created':   { icon: Plus,      color: 'text-blue-400',   label: 'Task created' },
			'task.updated':   { icon: CheckSquare, color: 'text-zinc-400', label: 'Task updated' },
			'task.deleted':   { icon: Trash2,    color: 'text-red-400',    label: 'Task deleted' },
			'task.assigned':  { icon: CheckSquare, color: 'text-blue-400', label: 'Task assigned' },
			'repo.created':   { icon: Zap,       color: 'text-yellow-400', label: 'Repo created' },
			'user.joined':    { icon: User,      color: 'text-green-400',  label: 'User joined' },
			'user.left':      { icon: User,      color: 'text-zinc-500',   label: 'User left' },
		};
		return map[type] ?? { icon: Zap, color: 'text-zinc-500', label: type };
	}

	function actorLabel(ev: TimelineEvent): string {
		if (ev.actor_type === 'agent') return 'Agent';
		if (ev.actor_type === 'system') return 'System';
		return ev.actor_id ? `User ${ev.actor_id.slice(-6)}` : 'User';
	}
</script>

<div class="flex h-full flex-col {className}">
	<div class="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
		<h2 class="text-sm font-semibold text-zinc-200">Event Timeline</h2>
		{#if events.loading}
			<div class="h-3 w-3 animate-spin rounded-full border-2 border-zinc-600 border-t-zinc-300"></div>
		{/if}
	</div>

	{#if error}
		<div class="m-3 flex items-center gap-2 rounded-md border border-red-900 bg-red-950/50 px-3 py-2 text-xs text-red-400">
			<AlertCircle class="h-3.5 w-3.5 shrink-0" />
			{error}
		</div>
	{/if}

	<div class="flex-1 overflow-y-auto">
		{#if repoEvents.length === 0 && !events.loading}
			<div class="flex flex-col items-center justify-center py-12 text-center">
				<Zap class="mb-2 h-8 w-8 text-zinc-700" />
				<p class="text-xs text-zinc-600">No events yet.</p>
			</div>
		{:else}
			<ol class="relative px-4 py-3">
				{#each repoEvents as ev, i (ev.id)}
					{@const meta = eventMeta(ev.type)}
					{@const Icon = meta.icon}
					<li class="relative mb-4 pl-6 last:mb-0">
						<!-- Timeline spine -->
						{#if i < repoEvents.length - 1}
							<span class="absolute left-[7px] top-5 bottom-[-1rem] w-px bg-zinc-800"></span>
						{/if}
						<!-- Dot -->
						<span class="absolute left-0 top-1 flex h-3.5 w-3.5 items-center justify-center rounded-full border border-zinc-700 bg-zinc-900">
							<Icon class="h-2 w-2 {meta.color}" />
						</span>

						<div class="flex items-start justify-between gap-2">
							<div class="min-w-0 flex-1">
								<p class="text-xs font-medium {meta.color}">{meta.label}</p>
								<p class="mt-0.5 text-[10px] text-zinc-600">{actorLabel(ev)}</p>
								{#if ev.payload && Object.keys(ev.payload).length > 0}
									<p class="mt-0.5 truncate font-mono text-[10px] text-zinc-700">
										{JSON.stringify(ev.payload).slice(0, 60)}
									</p>
								{/if}
							</div>
							<div class="shrink-0 text-right">
								<p class="text-[10px] text-zinc-600">{formatTime(ev.created_at)}</p>
								<p class="text-[10px] text-zinc-700">{formatDate(ev.created_at)}</p>
							</div>
						</div>
					</li>
				{/each}
			</ol>
		{/if}
	</div>
</div>
