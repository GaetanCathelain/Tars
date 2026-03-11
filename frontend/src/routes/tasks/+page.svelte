<script lang="ts">
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		// Redirect to first task if available
		if (tasksStore.tasks.length > 0) {
			goto(`/tasks/${tasksStore.tasks[0].id}`);
		}
	});
</script>

<div class="flex-1 flex flex-col items-center justify-center text-center px-6">
	{#if tasksStore.loading}
		<p class="text-text-secondary font-mono text-sm">Loading tasks...</p>
	{:else if tasksStore.tasks.length === 0}
		<div class="max-w-md">
			<div class="text-5xl mb-4">🤖</div>
			<h2 class="text-xl font-semibold text-text-primary mb-2">No tasks yet</h2>
			<p class="text-sm text-text-secondary mb-6">
				Create your first task to start orchestrating Claude Code sessions.
			</p>
			<p class="text-xs text-text-secondary font-mono">
				Use the "+ New Task" button in the sidebar to get started.
			</p>
		</div>
	{:else}
		<p class="text-text-secondary font-mono text-sm">Select a task from the sidebar</p>
	{/if}
</div>
