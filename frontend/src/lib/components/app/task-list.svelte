<script lang="ts">
	import { goto } from '$app/navigation';
	import { cn } from '$lib/utils';
	import { tasksStore } from '$lib/stores/tasks.svelte';

	function statusColor(status: string): string {
		switch (status) {
			case 'completed': return 'bg-emerald-500';
			case 'running': return 'bg-blue-500 animate-pulse';
			case 'failed': return 'bg-red-500';
			default: return 'bg-zinc-500';
		}
	}

	function handleSelect(id: string) {
		tasksStore.selectTask(id);
		goto(`/tasks/${id}`);
	}
</script>

<div class="space-y-1">
	{#each tasksStore.tasks as task (task.id)}
		<button
			class={cn(
				'flex items-center gap-3 px-3 py-2 rounded-md w-full text-left text-sm transition-colors',
				tasksStore.selectedTaskId === task.id
					? 'bg-accent text-accent-foreground'
					: 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground'
			)}
			onclick={() => handleSelect(task.id)}
		>
			<div class={cn('h-2 w-2 rounded-full shrink-0', statusColor(task.status))}></div>
			<span class="truncate">{task.title}</span>
		</button>
	{/each}
</div>
