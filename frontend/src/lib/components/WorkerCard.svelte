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

<div class="my-2 rounded-lg border border-border bg-bg-tertiary overflow-hidden shadow-[0_1px_2px_rgba(0,0,0,0.3)]">
	<!-- Header -->
	<button
		onclick={() => (collapsed = !collapsed)}
		class="w-full flex items-center justify-between px-3.5 py-2 hover:bg-bg-elevated/50 transition-all duration-150 cursor-pointer"
	>
		<div class="flex items-center gap-2">
			<span class="text-[11px] font-mono text-text-tertiary">Worker</span>
			<span class="text-[11px] font-mono text-text-secondary">{session.command || 'claude-code'}</span>

			{#if isRunning}
				<span class="inline-flex items-center gap-1.5 px-1.5 py-0.5 rounded text-[10px] font-medium text-running">
					<span class="w-1.5 h-1.5 rounded-full bg-running animate-pulse"></span>
					Running
				</span>
			{:else if isCompleted}
				<span class="inline-flex items-center gap-1.5 px-1.5 py-0.5 rounded text-[10px] font-medium text-success">
					<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
					</svg>
					{#if duration}Completed in {duration}{:else}Completed{/if}
				</span>
			{:else if isFailed}
				<span class="inline-flex items-center gap-1.5 px-1.5 py-0.5 rounded text-[10px] font-medium text-danger">
					<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
					{#if duration}Failed after {duration}{:else}Failed{/if}
				</span>
			{/if}

			{#if isFailed && session.exit_code !== undefined}
				<span class="text-[10px] font-mono text-danger">
					exit {session.exit_code}
				</span>
			{/if}
		</div>

		<svg
			class="w-3.5 h-3.5 text-text-tertiary transition-transform duration-200 {collapsed ? '' : 'rotate-180'}"
			fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
		</svg>
	</button>

	<!-- Terminal body -->
	{#if !collapsed}
		<div class="border-t border-border transition-all duration-200">
			{#if loaded}
				<Terminal
					bind:this={terminalRef}
					sessionId={session.id}
					{initialData}
				/>
			{:else}
				<div class="px-4 py-8 text-center">
					<p class="text-[11px] font-mono text-text-tertiary animate-pulse">Loading output...</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
