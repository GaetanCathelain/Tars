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
		<p class="text-text-tertiary text-[13px]">Loading tasks...</p>
	{:else if tasksStore.tasks.length === 0}
		<div class="max-w-sm">
			<div class="w-10 h-10 rounded-full bg-bg-elevated flex items-center justify-center mx-auto mb-4">
				<span class="text-text-tertiary text-lg">T</span>
			</div>
			<h2 class="text-[15px] font-semibold text-text-primary mb-1.5 tracking-[-0.01em]">No tasks yet</h2>
			<p class="text-[13px] text-text-secondary mb-4 leading-relaxed">
				Create your first task to start orchestrating Claude Code sessions.
			</p>
			<p class="text-[11px] text-text-tertiary">
				Use the "+ New Task" button in the sidebar to get started.
			</p>
		</div>
	{:else}
		<p class="text-text-tertiary text-[13px]">Select a task from the sidebar</p>
	{/if}
</div>
