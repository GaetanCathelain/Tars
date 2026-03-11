<script lang="ts">
	import type { WorkerSession } from '$lib/types';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import TerminalOutput from './terminal-output.svelte';

	let { worker }: { worker: WorkerSession } = $props();

	let expanded = $state(true);

	const isRunning = $derived(worker.status === 'running');

	function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'completed': return 'default';
			case 'running': return 'secondary';
			case 'failed': return 'destructive';
			default: return 'outline';
		}
	}

	function toggleExpanded() {
		expanded = !expanded;
	}
</script>

<Card.Root class="my-4 overflow-hidden">
	<button
		class="w-full py-3 px-4 flex items-center justify-between bg-muted/30 hover:bg-muted/50 transition-colors cursor-pointer border-none text-left"
		onclick={toggleExpanded}
	>
		<div class="flex items-center gap-2">
			{#if isRunning}
				<div class="h-2 w-2 rounded-full bg-primary animate-pulse"></div>
			{:else if worker.status === 'completed'}
				<span class="text-sm">✓</span>
			{:else}
				<span class="text-sm">✗</span>
			{/if}
			<Badge variant={statusVariant(worker.status)}>
				{worker.status}
			</Badge>
			<span class="text-xs text-muted-foreground font-mono ml-2 truncate">
				{worker.id.slice(0, 8)}
			</span>
		</div>
		<div class="flex items-center gap-2">
			<span class="text-xs text-muted-foreground font-mono">claude-code</span>
			<span class="text-xs text-muted-foreground">{expanded ? '▼' : '▶'}</span>
		</div>
	</button>

	{#if expanded}
		<TerminalOutput sessionId={worker.id} isLive={isRunning} />
	{/if}
</Card.Root>
