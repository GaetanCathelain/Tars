<script lang="ts">
	import type { PageData } from './$types';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import {
		Bot, CircleDot, GitFork, CheckSquare, Terminal,
		Square, ArrowRight, Activity, Zap
	} from 'lucide-svelte';
	import { cn } from '$lib/utils/cn';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	// Derived stats
	const runningAgents = $derived(data.agents.filter((a) => a.status === 'running' || a.status === 'starting'));
	const stoppedAgents = $derived(data.agents.filter((a) => a.status === 'stopped' || a.status === 'crashed'));
	const activeTasks = $derived(data.tasks.filter((t) => t.status === 'in_progress'));
	const pendingTasks = $derived(data.tasks.filter((t) => t.status === 'pending'));

	// Repo lookup map
	const repoMap = $derived(
		Object.fromEntries(data.repos.map((r) => [r.id, r]))
	);

	// Task lookup map
	const taskMap = $derived(
		Object.fromEntries(data.tasks.map((t) => [t.id, t]))
	);

	const statusColors: Record<string, string> = {
		running: 'border-green-900 bg-green-950/20',
		starting: 'border-yellow-900 bg-yellow-950/20',
		stopped: 'border-zinc-800 bg-zinc-900',
		crashed: 'border-red-900 bg-red-950/20'
	};

	const statusDotColors: Record<string, string> = {
		running: 'text-green-400',
		starting: 'text-yellow-400',
		stopped: 'text-zinc-600',
		crashed: 'text-red-400'
	};

	const statusLabels: Record<string, string> = {
		running: 'Running',
		starting: 'Starting',
		stopped: 'Stopped',
		crashed: 'Crashed'
	};

	const personaColors: Record<string, string> = {
		backend: 'text-blue-400 border-blue-900',
		frontend: 'text-purple-400 border-purple-900',
		devops: 'text-orange-400 border-orange-900',
		qa: 'text-green-400 border-green-900',
		general: 'text-zinc-400 border-zinc-700'
	};
</script>

<div class="flex h-full flex-col overflow-y-auto">
	<div class="p-6">
		<!-- Page header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold tracking-tight text-zinc-50">Overview</h1>
			<p class="mt-1 text-sm text-zinc-400">
				Welcome back{data.user?.name || data.user?.login ? `, ${data.user.name || data.user.login}` : ''}.
			</p>
		</div>

		<!-- Stats row -->
		<div class="mb-8 grid grid-cols-2 gap-3 sm:grid-cols-4">
			<a
				href="/repos"
				class="group rounded-xl border border-zinc-800 bg-zinc-900 p-4 transition-colors hover:border-zinc-700"
			>
				<div class="mb-2 flex items-center gap-2">
					<GitFork class="h-4 w-4 text-zinc-500" />
					<span class="text-xs font-medium uppercase tracking-wider text-zinc-500">Repos</span>
				</div>
				<p class="text-3xl font-bold tabular-nums text-zinc-50">{data.repos.length}</p>
			</a>

			<a
				href="/agents"
				class="group rounded-xl border border-green-900 bg-green-950/20 p-4 transition-colors hover:border-green-700"
			>
				<div class="mb-2 flex items-center gap-2">
					<Activity class="h-4 w-4 text-green-500" />
					<span class="text-xs font-medium uppercase tracking-wider text-green-600">Running</span>
				</div>
				<p class="text-3xl font-bold tabular-nums text-green-400">{runningAgents.length}</p>
			</a>

			<a
				href="/tasks"
				class="group rounded-xl border border-blue-900 bg-blue-950/20 p-4 transition-colors hover:border-blue-700"
			>
				<div class="mb-2 flex items-center gap-2">
					<Zap class="h-4 w-4 text-blue-500" />
					<span class="text-xs font-medium uppercase tracking-wider text-blue-600">In Progress</span>
				</div>
				<p class="text-3xl font-bold tabular-nums text-blue-400">{activeTasks.length}</p>
			</a>

			<a
				href="/tasks"
				class="group rounded-xl border border-zinc-800 bg-zinc-900 p-4 transition-colors hover:border-zinc-700"
			>
				<div class="mb-2 flex items-center gap-2">
					<CheckSquare class="h-4 w-4 text-zinc-500" />
					<span class="text-xs font-medium uppercase tracking-wider text-zinc-500">Pending</span>
				</div>
				<p class="text-3xl font-bold tabular-nums text-zinc-50">{pendingTasks.length}</p>
			</a>
		</div>

		<!-- Active agents section -->
		<div class="mb-8">
			<div class="mb-3 flex items-center justify-between">
				<h2 class="text-sm font-semibold text-zinc-300">Active Agents</h2>
				<a
					href="/agents"
					class="flex items-center gap-1 text-xs text-zinc-500 transition-colors hover:text-zinc-300"
				>
					View all <ArrowRight class="h-3 w-3" />
				</a>
			</div>

			{#if runningAgents.length === 0}
				<div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-zinc-800 py-10 text-center">
					<Bot class="mb-2 h-8 w-8 text-zinc-700" />
					<p class="text-sm text-zinc-500">No agents running</p>
					<a href="/agents" class="mt-2 text-xs text-zinc-600 hover:text-zinc-400">Go to Agents →</a>
				</div>
			{:else}
				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
					{#each runningAgents as agent (agent.id)}
						{@const repo = repoMap[agent.repo_id]}
						{@const task = agent.task_id ? taskMap[agent.task_id] : null}
						<div class={cn('flex flex-col gap-3 rounded-xl border p-4', statusColors[agent.status] ?? 'border-zinc-800 bg-zinc-900')}>
							<!-- Agent header -->
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-1.5">
										<CircleDot class="h-3.5 w-3.5 shrink-0 {statusDotColors[agent.status] ?? 'text-zinc-500'}" />
										<span class="truncate font-mono text-sm font-semibold text-zinc-50">
											{agent.name}
										</span>
									</div>
									{#if repo}
										<p class="mt-0.5 truncate text-xs text-zinc-600">{repo.name}</p>
									{/if}
								</div>
								<Badge
									variant="outline"
									class="shrink-0 text-[10px] font-normal {personaColors[agent.persona] ?? 'text-zinc-400 border-zinc-700'}"
								>
									{agent.persona}
								</Badge>
							</div>

							<!-- Task info -->
							{#if task}
								<div class="rounded-md border border-zinc-800 bg-zinc-950/50 px-2.5 py-2">
									<p class="text-[10px] uppercase tracking-wider text-zinc-600">Current task</p>
									<p class="mt-0.5 truncate text-xs text-zinc-300">{task.title}</p>
								</div>
							{/if}

							<!-- Branch -->
							<p class="truncate font-mono text-[10px] text-zinc-600">{agent.branch}</p>

							<!-- Status + actions -->
							<div class="flex items-center justify-between gap-2">
								<span class="text-xs {statusDotColors[agent.status] ?? 'text-zinc-500'}">
									{statusLabels[agent.status] ?? agent.status}
								</span>
								<div class="flex items-center gap-1.5">
									<form method="POST" action="/agents/{agent.id}/stop?repoId={agent.repo_id}">
										<Button type="submit" variant="outline" size="sm" class="h-6 gap-1 px-2 text-[10px]">
											<Square class="h-2.5 w-2.5 fill-current" />
											Stop
										</Button>
									</form>
									<a href="/agents/{agent.id}?repoId={agent.repo_id}">
										<Button variant="ghost" size="sm" class="h-6 gap-1 px-2 text-[10px]">
											<Terminal class="h-2.5 w-2.5" />
											Terminal
										</Button>
									</a>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Stopped / crashed agents -->
		{#if stoppedAgents.length > 0}
			<div class="mb-8">
				<div class="mb-3 flex items-center justify-between">
					<h2 class="text-sm font-semibold text-zinc-500">Stopped Agents</h2>
				</div>
				<div class="flex flex-col divide-y divide-zinc-800 rounded-xl border border-zinc-800 overflow-hidden">
					{#each stoppedAgents as agent (agent.id)}
						{@const repo = repoMap[agent.repo_id]}
						<a
							href="/agents/{agent.id}?repoId={agent.repo_id}"
							class="flex items-center gap-3 px-4 py-3 transition-colors hover:bg-zinc-800/40"
						>
							<CircleDot class="h-3.5 w-3.5 shrink-0 {statusDotColors[agent.status] ?? 'text-zinc-600'}" />
							<span class="min-w-0 flex-1 truncate font-mono text-sm text-zinc-400">{agent.name}</span>
							{#if repo}
								<span class="shrink-0 text-xs text-zinc-600">{repo.name}</span>
							{/if}
							<Badge variant="outline" class="shrink-0 text-[10px] {agent.status === 'crashed' ? 'text-red-400 border-red-900' : 'text-zinc-600 border-zinc-800'}">
								{agent.status}
							</Badge>
						</a>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Quick links -->
		{#if data.repos.length > 0 && runningAgents.length === 0 && data.agents.length === 0}
			<div class="rounded-xl border border-zinc-800 bg-zinc-900 p-5">
				<h3 class="mb-3 text-sm font-semibold text-zinc-300">Get started</h3>
				<div class="flex flex-col gap-2 text-sm text-zinc-500">
					<a href="/repos" class="flex items-center gap-2 hover:text-zinc-200">
						<GitFork class="h-4 w-4" /> Configure a repository
					</a>
					<a href="/tasks" class="flex items-center gap-2 hover:text-zinc-200">
						<CheckSquare class="h-4 w-4" /> Create tasks on the board
					</a>
					<a href="/agents" class="flex items-center gap-2 hover:text-zinc-200">
						<Bot class="h-4 w-4" /> Spawn an agent
					</a>
				</div>
			</div>
		{/if}
	</div>
</div>
