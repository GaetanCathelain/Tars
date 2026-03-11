<script lang="ts">
	import '../app.css';
	import { authStore } from '$lib/stores/auth.svelte';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { wsStore } from '$lib/stores/websocket.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	let { children } = $props();

	let newTaskInput = $state('');
	let showNewTask = $state(false);

	$effect(() => {
		if (authStore.isAuthenticated) {
			tasksStore.fetchTasks();
			wsStore.connect();
		} else {
			wsStore.disconnect();
		}
	});

	const connectionDotClass = $derived.by(() => {
		switch (wsStore.status) {
			case 'connected': return 'bg-success';
			case 'connecting': return 'bg-warning animate-pulse';
			case 'disconnected': return 'bg-danger';
			default: return 'bg-text-secondary';
		}
	});

	function statusColor(status: string): string {
		switch (status) {
			case 'open':
			case 'running':
				return 'bg-success';
			case 'completed':
				return 'bg-warning';
			case 'failed':
				return 'bg-danger';
			default:
				return 'bg-text-secondary';
		}
	}

	async function handleCreateTask() {
		if (!newTaskInput.trim()) return;
		const task = await tasksStore.createTask(newTaskInput.trim());
		if (task) {
			newTaskInput = '';
			showNewTask = false;
			goto(`/tasks/${task.id}`);
		}
	}

	function handleTaskKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleCreateTask();
		if (e.key === 'Escape') {
			showNewTask = false;
			newTaskInput = '';
		}
	}
</script>

{#if !authStore.isAuthenticated}
	<div class="min-h-screen flex items-center justify-center bg-bg-primary">
		{@render children()}
	</div>
{:else}
	<div class="flex h-screen bg-bg-primary">
		<!-- Sidebar -->
		<aside class="w-72 flex flex-col bg-bg-secondary border-r border-border shrink-0">
			<!-- App title -->
			<div class="px-5 py-4 border-b border-border">
				<h1 class="text-lg font-mono font-bold tracking-wider text-accent">TARS</h1>
				<p class="text-xs text-text-secondary mt-0.5">Orchestrator</p>
			</div>

			<!-- Task list -->
			<nav class="flex-1 overflow-y-auto py-2">
				{#each tasksStore.tasks as task}
					{@const isActive = page.url.pathname === `/tasks/${task.id}`}
					<a
						href="/tasks/{task.id}"
						class="flex items-center gap-3 px-5 py-3 text-sm transition-colors
							{isActive
								? 'bg-bg-tertiary text-text-primary'
								: 'text-text-secondary hover:bg-bg-tertiary/50 hover:text-text-primary'}"
					>
						<span class="w-2 h-2 rounded-full shrink-0 {statusColor(task.status)}"></span>
						<span class="truncate">{task.title}</span>
					</a>
				{/each}

				{#if tasksStore.tasks.length === 0 && !tasksStore.loading}
					<p class="px-5 py-8 text-sm text-text-secondary text-center">No tasks yet</p>
				{/if}
			</nav>

			<!-- New Task / User -->
			<div class="border-t border-border p-3">
				{#if showNewTask}
					<div class="flex gap-2">
						<input
							type="text"
							bind:value={newTaskInput}
							onkeydown={handleTaskKeydown}
							placeholder="Task description..."
							class="flex-1 bg-bg-tertiary border border-border rounded px-3 py-2 text-sm text-text-primary
								placeholder:text-text-secondary focus:outline-none focus:border-accent transition-colors"
						/>
						<button
							onclick={handleCreateTask}
							class="px-3 py-2 bg-accent text-bg-primary text-sm font-medium rounded
								hover:bg-accent-hover transition-colors"
						>
							+
						</button>
					</div>
				{:else}
					<button
						onclick={() => (showNewTask = true)}
						class="w-full py-2 text-sm text-text-secondary border border-border rounded
							hover:border-accent hover:text-accent transition-colors"
					>
						+ New Task
					</button>
				{/if}

				<div class="flex items-center justify-between mt-3 px-1">
					<div class="flex items-center gap-2">
						<span class="w-2 h-2 rounded-full {connectionDotClass}" title="WebSocket: {wsStore.status}"></span>
						<span class="text-xs text-text-secondary font-mono">
							{authStore.user?.username}
						</span>
					</div>
					<button
						onclick={() => { authStore.logout(); goto('/login'); }}
						class="text-xs text-text-secondary hover:text-danger transition-colors"
					>
						Logout
					</button>
				</div>
			</div>
		</aside>

		<!-- Main content -->
		<main class="flex-1 flex flex-col min-w-0">
			{@render children()}
		</main>
	</div>
{/if}
