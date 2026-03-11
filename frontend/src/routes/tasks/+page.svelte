<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import Sidebar from '$lib/components/app/sidebar.svelte';

	onMount(() => {
		auth.init();
		if (!auth.isAuthenticated) {
			goto('/login');
			return;
		}
		tasksStore.fetchTasks();
	});
</script>

<div class="flex h-screen bg-background">
	<Sidebar />
	<main class="flex-1 flex items-center justify-center">
		<div class="text-center space-y-2">
			<p class="text-lg text-muted-foreground">Select a task to get started</p>
			<p class="text-sm text-muted-foreground/60">or create a new one from the sidebar</p>
		</div>
	</main>
</div>
