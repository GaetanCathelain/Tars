<script lang="ts">
	import type { WorkerSession } from '$lib/types';
	import * as Card from '$lib/components/ui/card';

	let { worker }: { worker: WorkerSession } = $props();

	const isRunning = $derived(worker.status === 'running');
</script>

<Card.Root class="my-4 overflow-hidden">
	<Card.Header class="py-3 px-4 flex flex-row items-center justify-between space-y-0 bg-muted/30">
		<div class="flex items-center gap-2">
			{#if isRunning}
				<div class="h-2 w-2 rounded-full bg-primary animate-pulse"></div>
				<span class="text-sm font-medium">Running</span>
			{:else if worker.status === 'completed'}
				<span class="text-sm font-medium text-muted-foreground">✓ Completed</span>
			{:else}
				<span class="text-sm font-medium text-destructive">✗ Failed</span>
			{/if}
		</div>
		<span class="text-xs text-muted-foreground font-mono">worker: claude-code</span>
	</Card.Header>
	<div class="bg-black min-h-[200px] max-h-[400px] flex items-center justify-center">
		<p class="text-zinc-600 text-sm font-mono">Terminal output placeholder</p>
	</div>
</Card.Root>
