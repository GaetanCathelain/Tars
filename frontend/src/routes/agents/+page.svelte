<script lang="ts">
	import type { PageData } from './$types';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import SpawnAgentDialog from '$lib/components/SpawnAgentDialog.svelte';
	import { Bot, CircleDot, Terminal, Plus } from 'lucide-svelte';
	import { cn } from '$lib/utils/cn';
	import { goto } from '$app/navigation';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	const selectedRepoId = $derived(data.selectedRepoId);
	const agents = $derived(data.agents);
	const tasks = $derived(data.tasks ?? []);

	let dialogOpen = $state(false);

	const statusColors: Record<string, string> = {
		running: 'text-green-400 border-green-900 bg-green-950/30',
		starting: 'text-yellow-400 border-yellow-900 bg-yellow-950/30',
		stopped: 'text-zinc-500 border-zinc-800 bg-zinc-900',
		crashed: 'text-red-400 border-red-900 bg-red-950/30'
	};

	const statusDotColors: Record<string, string> = {
		running: 'text-green-400',
		starting: 'text-yellow-400',
		stopped: 'text-zinc-600',
		crashed: 'text-red-400'
	};

	function handleSpawnSuccess() {
		// Reload to show the new agent
		goto(`/agents?repoId=${selectedRepoId}`, { invalidateAll: true });
	}
</script>

<div class="p-6">
	<!-- Header + repo selector -->
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-zinc-50">Agents</h1>
			<p class="mt-1 text-sm text-zinc-400">
				{agents.length} agent{agents.length !== 1 ? 's' : ''}
			</p>
		</div>

		<div class="flex items-center gap-3">
			{#if data.repos.length > 1}
				<select
					onchange={(e) => {
						const id = (e.target as HTMLSelectElement).value;
						window.location.href = `/agents?repoId=${id}`;
					}}
					class="rounded-md border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
				>
					{#each data.repos as repo (repo.id)}
						<option value={repo.id} selected={repo.id === selectedRepoId}>
							{repo.name}
						</option>
					{/each}
				</select>
			{/if}

			{#if selectedRepoId}
				<Button
					size="sm"
					class="gap-1.5"
					onclick={() => (dialogOpen = true)}
				>
					<Plus class="h-3.5 w-3.5" />
					Spawn Agent
				</Button>
			{/if}
		</div>
	</div>

	<!-- Agent list -->
	{#if agents.length === 0}
		<div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-zinc-800 py-16 text-center">
			<Bot class="mb-3 h-10 w-10 text-zinc-700" />
			<p class="text-sm font-medium text-zinc-400">No agents running</p>
			<p class="mt-1 text-xs text-zinc-600">Spawn an agent to get started.</p>
			{#if selectedRepoId}
				<Button size="sm" class="mt-4 gap-1.5" onclick={() => (dialogOpen = true)}>
					<Plus class="h-3.5 w-3.5" />
					Spawn your first agent
				</Button>
			{/if}
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each agents as agent (agent.id)}
				<a
					href="/agents/{agent.id}?repoId={selectedRepoId}"
					class={cn(
						'group flex flex-col gap-3 rounded-xl border p-4 transition-colors hover:border-zinc-600',
						statusColors[agent.status] ?? 'border-zinc-800 bg-zinc-900'
					)}
				>
					<div class="flex items-start justify-between gap-2">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<CircleDot class="h-3.5 w-3.5 shrink-0 {statusDotColors[agent.status] ?? 'text-zinc-500'}" />
								<span class="truncate font-mono text-sm font-semibold text-zinc-50">
									{agent.name}
								</span>
							</div>
							<p class="mt-0.5 truncate font-mono text-xs text-zinc-500">{agent.branch}</p>
						</div>
						<Badge variant="outline" class="shrink-0 text-xs font-normal">
							{agent.persona}
						</Badge>
					</div>

					<div class="flex items-center justify-between text-xs text-zinc-500">
						<span>{agent.status}</span>
						<span class="flex items-center gap-1 text-zinc-600 transition-colors group-hover:text-zinc-400">
							<Terminal class="h-3 w-3" />
							Open terminal
						</span>
					</div>

					{#if agent.task_id}
						<p class="truncate text-xs text-zinc-600">Task: {agent.task_id}</p>
					{/if}
				</a>
			{/each}
		</div>
	{/if}
</div>

<!-- Spawn dialog -->
{#if selectedRepoId}
	<SpawnAgentDialog
		repoId={selectedRepoId}
		{tasks}
		bind:open={dialogOpen}
		onsuccess={handleSpawnSuccess}
	/>
{/if}
