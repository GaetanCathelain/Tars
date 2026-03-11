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
				return 'bg-running animate-pulse';
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
	<div class="min-h-screen flex items-center justify-center bg-[#111113]">
		{@render children()}
	</div>
{:else}
	<div class="flex h-screen bg-bg-primary">
		<!-- Sidebar -->
		<aside class="w-64 flex flex-col bg-bg-secondary border-r border-zinc-800/50 shrink-0">
			<!-- App title -->
			<div class="px-5 py-5 border-b border-zinc-800/50">
				<h1 class="text-base font-semibold tracking-wide text-zinc-100">TARS</h1>
				<p class="text-xs text-zinc-500 mt-0.5">Orchestrator</p>
			</div>

			<!-- Section label -->
			<div class="px-5 pt-4 pb-2">
				<span class="text-[11px] font-semibold uppercase tracking-[0.1em] text-zinc-500">Tasks</span>
			</div>

			<!-- Task list -->
			<nav class="flex-1 overflow-y-auto px-2">
				{#each tasksStore.tasks as task}
					{@const isActive = page.url.pathname === `/tasks/${task.id}`}
					<a
						href="/tasks/{task.id}"
						class="flex items-center gap-3 px-3 py-2.5 mx-0 rounded-lg text-sm transition-colors duration-150
							{isActive
								? 'bg-indigo-500/10 text-zinc-100'
								: 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'}"
					>
						<span class="w-2 h-2 rounded-full shrink-0 {statusColor(task.status)}"></span>
						<span class="truncate">{task.title}</span>
					</a>
				{/each}

				{#if tasksStore.tasks.length === 0 && !tasksStore.loading}
					<p class="px-3 py-8 text-sm text-zinc-500 text-center">No tasks yet</p>
				{/if}
			</nav>

			<!-- New Task / User -->
			<div class="border-t border-zinc-800/50 p-3">
				{#if showNewTask}
					<div class="flex gap-2">
						<input
							type="text"
							bind:value={newTaskInput}
							onkeydown={handleTaskKeydown}
							placeholder="Task description..."
							class="flex-1 h-10 bg-[#1c1c20] border border-zinc-800 rounded-lg px-3 text-sm text-text-primary
								placeholder:text-text-tertiary focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
						/>
						<button
							onclick={handleCreateTask}
							class="h-10 px-3 bg-indigo-500 text-white text-sm font-medium rounded-lg
								hover:bg-indigo-600 transition-colors duration-150"
						>
							+
						</button>
					</div>
				{:else}
					<button
						onclick={() => (showNewTask = true)}
						class="w-full h-10 text-sm text-zinc-400 border border-dashed border-zinc-700 rounded-lg
							hover:border-zinc-500 hover:text-zinc-300 transition-colors duration-150"
					>
						+ New Task
					</button>
				{/if}

				<div class="flex items-center justify-between mt-3 px-2">
					<div class="flex items-center gap-2.5">
						<span class="w-2 h-2 rounded-full {connectionDotClass}" title="WebSocket: {wsStore.status}"></span>
						<span class="text-sm text-zinc-400">
							{authStore.user?.username}
						</span>
					</div>
					<button
						onclick={() => { authStore.logout(); goto('/login'); }}
						class="text-sm text-zinc-500 hover:text-danger transition-colors duration-150"
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
