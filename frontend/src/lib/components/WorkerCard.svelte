<script lang="ts">
	import { onMount } from 'svelte';
	import type { WorkerSession } from '$lib/types';
	import { wsStore } from '$lib/stores/websocket.svelte';
	import { workersStore } from '$lib/stores/workers.svelte';
	import Terminal from './Terminal.svelte';

	let {
		session
	}: {
		session: WorkerSession;
	} = $props();

	let collapsed = $state(false);
	let terminalRef: Terminal | undefined = $state();
	let initialData = $state<Uint8Array[]>([]);
	let loaded = $state(false);

	const isRunning = $derived(session.status === 'running');
	const isCompleted = $derived(session.status === 'completed');
	const isFailed = $derived(session.status === 'failed');

	const duration = $derived.by(() => {
		if (!session.started_at) return '';
		const start = new Date(session.started_at).getTime();
		const end = session.finished_at ? new Date(session.finished_at).getTime() : Date.now();
		const seconds = Math.floor((end - start) / 1000);
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		if (mins > 0) return `${mins}m ${secs}s`;
		return `${secs}s`;
	});

	onMount(() => {
		// If worker has history (completed/failed), fetch and replay
		if (!isRunning && session.id && !session.id.startsWith('mock-')) {
			workersStore.fetchOutput(session.id).then((chunks) => {
				initialData = chunks;
				loaded = true;
			});
		} else {
			loaded = true;
		}

		// Subscribe to live output if running
		let unsub: (() => void) | undefined;
		if (isRunning) {
			unsub = wsStore.onWorkerOutput(session.id, (_sid, data) => {
				if (terminalRef) {
					terminalRef.write(data);
				}
			});
		}

		return () => {
			if (unsub) unsub();
		};
	});
</script>

<div class="my-2 bg-zinc-900/50 border border-zinc-800 rounded-xl shadow-md overflow-hidden">
	<!-- Header -->
	<button
		onclick={() => (collapsed = !collapsed)}
		class="w-full flex items-center justify-between px-4 py-3 border-b border-zinc-800/50 hover:bg-zinc-800/30 transition-colors duration-150 cursor-pointer"
	>
		<div class="flex items-center gap-3">
			<span class="text-sm text-zinc-500">Worker:</span>
			<span class="text-sm font-medium text-zinc-300">{session.command || 'claude-code'}</span>

			{#if isRunning}
				<span class="inline-flex items-center gap-2 text-sm text-indigo-400">
					<span class="w-2 h-2 rounded-full bg-indigo-400 animate-pulse"></span>
					Running
				</span>
			{:else if isCompleted}
				<span class="inline-flex items-center gap-2 text-sm text-emerald-400">
					<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
					</svg>
					{#if duration}Completed in {duration}{:else}Completed{/if}
				</span>
			{:else if isFailed}
				<span class="inline-flex items-center gap-2 text-sm text-red-400">
					<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
					{#if duration}Failed after {duration}{:else}Failed{/if}
				</span>
			{/if}

			{#if isFailed && session.exit_code !== undefined}
				<span class="text-xs font-mono text-red-400">
					exit {session.exit_code}
				</span>
			{/if}
		</div>

		<svg
			class="w-4 h-4 text-zinc-500 transition-transform duration-200 {collapsed ? '' : 'rotate-180'}"
			fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
		</svg>
	</button>

	<!-- Terminal body -->
	{#if !collapsed}
		<div class="transition-all duration-200">
			{#if loaded}
				<Terminal
					bind:this={terminalRef}
					sessionId={session.id}
					{initialData}
				/>
			{:else}
				<div class="px-4 py-10 text-center">
					<p class="text-sm text-zinc-500 animate-pulse">Loading output...</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
