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
			default: return 'bg-text-tertiary';
		}
	});

	function statusColor(status: string): string {
		switch (status) {
			case 'open':
				return 'bg-text-tertiary';
			case 'running':
				return 'bg-running';
			case 'completed':
				return 'bg-success';
			case 'failed':
				return 'bg-danger';
			default:
				return 'bg-text-tertiary';
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
		<aside class="w-64 flex flex-col bg-bg-secondary border-r border-border shrink-0">
			<!-- App title -->
			<div class="px-4 py-4">
				<h1 class="text-sm font-semibold tracking-wide text-zinc-200">TARS</h1>
			</div>

			<!-- Section label -->
			<div class="px-4 pt-2 pb-1">
				<span class="text-[10px] font-semibold uppercase tracking-widest text-text-tertiary">Tasks</span>
			</div>

			<!-- Task list -->
			<nav class="flex-1 overflow-y-auto px-2">
				{#each tasksStore.tasks as task}
					{@const isActive = page.url.pathname === `/tasks/${task.id}`}
					<a
						href="/tasks/{task.id}"
						class="flex items-center gap-2.5 px-3 py-1.5 rounded-md text-[13px] transition-all duration-150
							{isActive
								? 'bg-accent-muted text-text-primary'
								: 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'}"
					>
						<span class="w-1.5 h-1.5 rounded-full shrink-0 {statusColor(task.status)}"></span>
						<span class="truncate">{task.title}</span>
					</a>
				{/each}

				{#if tasksStore.tasks.length === 0 && !tasksStore.loading}
					<p class="px-3 py-6 text-[13px] text-text-tertiary text-center">No tasks yet</p>
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
							class="flex-1 bg-bg-primary border border-border rounded-md px-3 py-1.5 text-[13px] text-text-primary
								placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
						/>
						<button
							onclick={handleCreateTask}
							class="px-2.5 py-1.5 bg-accent text-white text-[13px] font-medium rounded-md
								hover:bg-accent-hover transition-all duration-150"
						>
							+
						</button>
					</div>
				{:else}
					<button
						onclick={() => (showNewTask = true)}
						class="w-full py-1.5 text-[13px] text-text-tertiary border border-border rounded-md
							hover:border-text-tertiary hover:text-text-secondary transition-all duration-150"
					>
						+ New Task
					</button>
				{/if}

				<div class="flex items-center justify-between mt-3 px-1">
					<div class="flex items-center gap-2">
						<span class="w-1.5 h-1.5 rounded-full {connectionDotClass}" title="WebSocket: {wsStore.status}"></span>
						<span class="text-[11px] text-text-tertiary font-mono">
							{authStore.user?.username}
						</span>
					</div>
					<button
						onclick={() => { authStore.logout(); goto('/login'); }}
						class="text-[11px] text-text-tertiary hover:text-danger transition-all duration-150"
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
