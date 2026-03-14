<script lang="ts">
	import type { PageData } from './$types';
	import { onDestroy } from 'svelte';
	import { wsClient } from '$lib/ws/client.svelte';
	import { agents } from '$lib/stores/agents.svelte';
	import { presence } from '$lib/stores/presence.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import AgentTerminal from '$lib/components/AgentTerminal.svelte';
	import AgentChat from '$lib/components/AgentChat.svelte';
	import PresenceIndicators from '$lib/components/PresenceIndicators.svelte';
	import DiffViewer from '$lib/components/DiffViewer.svelte';
	import MergePanel from '$lib/components/MergePanel.svelte';
	import { ArrowLeft, CircleDot, Square, Terminal, MessageSquare, GitBranch, GitMerge } from 'lucide-svelte';
	import type { AgentLogLine, RepoDiff } from '$shared/types/models';
	import { cn } from '$lib/utils/cn';
	import { api } from '$lib/utils/api';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	type Tab = 'terminal' | 'chat' | 'diff' | 'merge';
	let activeTab = $state<Tab>('terminal');

	const agentId = $derived(data.agent.id);
	const repoId = $derived(data.repoId);
	const agentChannel = $derived(`agent:${agentId}`);
	const repoChannel = $derived(`repo:${repoId}`);
	const initialPresence = $derived(data.presence);
	const initialDiff = $derived(data.diff);

	// Seed presence from SSR data
	$effect(() => {
		if (initialPresence) {
			presence.setSnapshot(repoId, initialPresence.users);
		}
	});

	$effect(() => {
		agents.addAgent(data.agent);
	});

	$effect(() => {
		wsClient.subscribe(agentChannel);
		wsClient.subscribe(repoChannel);
		wsClient.presenceUpdate({ repo_id: repoId, viewing_agent_id: agentId });

		return () => {
			wsClient.unsubscribe(agentChannel);
			wsClient.presenceUpdate({ repo_id: repoId, viewing_agent_id: null });
		};
	});

	onDestroy(() => {
		wsClient.unsubscribe(agentChannel);
	});

	const liveAgent = $derived(
		agents.agents.find((a) => a.id === agentId) ?? { ...data.agent, outputLines: [] as AgentLogLine[] }
	);

	const statusColors: Record<string, string> = {
		running: 'text-green-400',
		starting: 'text-yellow-400',
		stopped: 'text-zinc-500',
		crashed: 'text-red-400'
	};

	function statusColor(s: string): string {
		return statusColors[s] ?? 'text-zinc-400';
	}

	const displayLines = $derived(
		liveAgent.outputLines.length > 0 ? liveAgent.outputLines : data.initialLines
	);

	// Live diff: refresh on tab switch to 'diff' / 'merge'
	let liveDiff = $state<RepoDiff | null>(null);
	$effect(() => { liveDiff = initialDiff ?? null; });

	async function refreshDiff() {
		try {
			liveDiff = await api.get<RepoDiff>(`/repos/${repoId}/agents/${agentId}/diff`);
		} catch {
			// Keep stale diff
		}
	}

	function switchTab(tab: Tab) {
		activeTab = tab;
		if (tab === 'diff' || tab === 'merge') {
			refreshDiff();
		}
	}

	const tabs: { id: Tab; label: string; icon: typeof Terminal }[] = [
		{ id: 'terminal', label: 'Terminal', icon: Terminal },
		{ id: 'chat', label: 'Chat', icon: MessageSquare },
		{ id: 'diff', label: 'Diff', icon: GitBranch },
		{ id: 'merge', label: 'Merge', icon: GitMerge }
	];
</script>

<div class="flex h-full flex-col">
	<!-- Header -->
	<div class="flex flex-wrap items-center gap-2 border-b border-zinc-800 bg-zinc-900 px-3 py-2 sm:gap-4 sm:px-5 sm:py-3">
		<a
			href="/agents?repoId={repoId}"
			class="flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-200"
		>
			<ArrowLeft class="h-4 w-4" />
			<span class="hidden sm:inline">Agents</span>
		</a>

		<div class="hidden h-4 w-px bg-zinc-700 sm:block"></div>

		<div class="flex min-w-0 flex-1 items-center gap-2 sm:gap-3">
			<span class="truncate font-mono text-sm font-semibold text-zinc-50">{liveAgent.name}</span>
			<Badge variant="outline" class="hidden shrink-0 text-xs font-normal text-zinc-400 sm:inline-flex">
				{liveAgent.persona}
			</Badge>
			<div class="flex shrink-0 items-center gap-1">
				<CircleDot class="h-3.5 w-3.5 {statusColor(liveAgent.status)}" />
				<span class="hidden text-xs sm:inline {statusColor(liveAgent.status)}">{liveAgent.status}</span>
			</div>
		</div>

		<div class="flex shrink-0 items-center gap-2 sm:gap-3">
			<!-- Presence indicators -->
			<PresenceIndicators {repoId} />

			<span class="hidden font-mono text-xs text-zinc-600 lg:inline">{liveAgent.branch}</span>
			{#if liveAgent.status === 'running' || liveAgent.status === 'starting'}
				<form method="POST" action="/agents/{agentId}/stop?repoId={repoId}">
					<Button type="submit" variant="destructive" size="sm" class="h-7 gap-1.5 text-xs">
						<Square class="h-3 w-3 fill-current" />
						Stop
					</Button>
				</form>
			{/if}
		</div>
	</div>

	<!-- Tabs — scrollable on small screens -->
	<div class="flex overflow-x-auto border-b border-zinc-800 bg-zinc-900/50">
		{#each tabs as tab (tab.id)}
			{@const Icon = tab.icon}
			<button
				onclick={() => switchTab(tab.id)}
				class={cn(
					'flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-2.5 text-sm font-medium transition-colors sm:px-4',
					activeTab === tab.id
						? 'border-zinc-50 text-zinc-50'
						: 'border-transparent text-zinc-500 hover:text-zinc-300'
				)}
			>
				<Icon class="h-3.5 w-3.5" />
				{tab.label}
			</button>
		{/each}
	</div>

	<!-- Content panels -->
	<div class="relative flex-1 overflow-hidden">
		<!-- Terminal -->
		<div class={cn('absolute inset-0 p-2', activeTab === 'terminal' ? 'block' : 'hidden')}>
			<AgentTerminal agentId={agentId} lines={displayLines} class="h-full" />
		</div>

		<!-- Chat -->
		<div class={cn('absolute inset-0', activeTab === 'chat' ? 'block' : 'hidden')}>
			<AgentChat
				agentId={agentId}
				repoId={repoId}
				agentName={liveAgent.name}
				agentStatus={liveAgent.status}
				outputLines={displayLines}
			/>
		</div>

		<!-- Diff -->
		{#if activeTab === 'diff'}
			<div class="absolute inset-0 overflow-y-auto p-4">
				{#if liveDiff}
					<DiffViewer diff={liveDiff} />
				{:else}
					<div class="flex h-full items-center justify-center">
						<p class="text-sm text-zinc-600">No diff available.</p>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Merge -->
		{#if activeTab === 'merge'}
			<div class="absolute inset-0">
				<MergePanel agent={liveAgent} repoId={repoId} diff={liveDiff} />
			</div>
		{/if}
	</div>
</div>
