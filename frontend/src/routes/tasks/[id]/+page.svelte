<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { messagesStore } from '$lib/stores/messages.svelte';
	import { workersStore } from '$lib/stores/workers.svelte';
	import Sidebar from '$lib/components/app/sidebar.svelte';
	import ChatView from '$lib/components/app/chat-view.svelte';

	let { data } = $props();

	onMount(() => {
		auth.init();
		if (!auth.isAuthenticated) {
			goto('/login');
			return;
		}
		tasksStore.fetchTasks().then(() => {
			tasksStore.selectTask(data.taskId);
			messagesStore.fetchMessages(data.taskId);
			workersStore.fetchWorkers(data.taskId);
		});
	});
</script>

<div class="flex h-screen bg-background">
	<Sidebar />
	<ChatView />
</div>
