<script lang="ts">
	import type { WorkerSession } from '$lib/types';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import TerminalOutput from './terminal-output.svelte';

	let { worker }: { worker: WorkerSession } = $props();

	let expanded = $state(false);
	let commandExpanded = $state(false);

	const isRunning = $derived(worker.status === 'running');

	// Extract a short title from the command — first line or first ~80 chars
	const commandTitle = $derived(() => {
		if (!worker.command) return 'Worker task';
		// Strip the leading "claude " prefix if present
		const prompt = worker.command.replace(/^claude\s+["']?/, '').replace(/["']$/, '');
		const firstLine = prompt.split('\n')[0];
		return firstLine.length > 80 ? firstLine.slice(0, 77) + '…' : firstLine;
	});

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

	function toggleCommand(e: MouseEvent) {
		e.stopPropagation();
		commandExpanded = !commandExpanded;
	}
</script>

<Card.Root class="my-4 overflow-hidden">
	<button
		class="w-full py-3 px-4 flex flex-col gap-1.5 bg-muted/30 hover:bg-muted/50 transition-colors cursor-pointer border-none text-left"
		onclick={toggleExpanded}
	>
		<div class="flex items-center justify-between w-full">
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
			</div>
			<div class="flex items-center gap-2">
				<span class="text-xs text-muted-foreground font-mono">claude-code</span>
				<span class="text-xs text-muted-foreground">{expanded ? '▼' : '▶'}</span>
			</div>
		</div>
		<p class="text-sm text-foreground/80 truncate w-full">{commandTitle()}</p>
	</button>

	{#if worker.command}
		<div class="px-4 py-1.5 bg-muted/20 border-t border-border/50 flex items-center gap-1">
			<button
				class="text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer border-none bg-transparent flex items-center gap-1"
				onclick={toggleCommand}
			>
				<span>{commandExpanded ? '▾' : '▸'}</span>
				<span>Full command</span>
			</button>
		</div>
		{#if commandExpanded}
			<div class="px-4 py-3 bg-muted/10 border-t border-border/30 max-h-[300px] overflow-y-auto">
				<pre class="text-xs text-muted-foreground font-mono whitespace-pre-wrap break-words">{worker.command}</pre>
			</div>
		{/if}
	{/if}

	{#if expanded}
		<TerminalOutput sessionId={worker.id} isLive={isRunning} />
	{/if}
</Card.Root>
