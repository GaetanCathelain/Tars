<script lang="ts">
	import { onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { wsClient } from '$lib/ws/client.svelte';
	import { tasks } from '$lib/stores/tasks.svelte';
	import { api, ApiError } from '$lib/utils/api';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Plus, GripVertical, AlertCircle, Loader2 } from 'lucide-svelte';
	import type { Task, TaskStatus, Repo } from '$shared/types/models';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	// Repo selector
	const repos: Repo[] = $derived(data.repos ?? []);
	const selectedRepoId: string | null = $derived(data.selectedRepoId ?? null);

	// Column definitions
	const COLUMNS: { status: TaskStatus; label: string; color: string }[] = [
		{ status: 'pending', label: 'Pending', color: 'text-zinc-400' },
		{ status: 'in_progress', label: 'In Progress', color: 'text-blue-400' },
		{ status: 'done', label: 'Done', color: 'text-green-400' },
		{ status: 'cancelled', label: 'Cancelled', color: 'text-zinc-600' }
	];

	// Local state
	let error = $state<string | null>(null);
	let newTaskTitle = $state('');
	let addingToColumn = $state<TaskStatus | null>(null);
	let submitting = $state(false);
	let draggingTask = $state<Task | null>(null);
	let dragOverColumn = $state<TaskStatus | null>(null);

	// Seed the store from server-side data
	$effect(() => {
		tasks.setTasks(data.tasks ?? []);
	});

	// Subscribe to repo WS channel when repoId is available
	$effect(() => {
		if (!selectedRepoId) return;
		const channel = `repo:${selectedRepoId}`;
		wsClient.subscribe(channel);

		const off = wsClient.onMessage((msg) => {
			if (msg.type === 'task.created') tasks.addTask(msg.payload.task);
			else if (msg.type === 'task.updated') tasks.updateTask(msg.payload.task);
			else if (msg.type === 'task.deleted') tasks.removeTask(msg.payload.task_id);
		});

		return () => {
			wsClient.unsubscribe(channel);
			off();
		};
	});

	onDestroy(() => {
		if (selectedRepoId) wsClient.unsubscribe(`repo:${selectedRepoId}`);
	});

	// Derived: tasks grouped by status
	const byStatus = $derived.by(() => {
		const map: Record<TaskStatus, Task[]> = {
			pending: [],
			in_progress: [],
			done: [],
			cancelled: []
		};
		for (const t of tasks.tasks) {
			const col = map[t.status as TaskStatus];
			if (col) col.push(t);
		}
		return map;
	});

	async function createTask(status: TaskStatus) {
		const title = newTaskTitle.trim();
		if (!title || !selectedRepoId || submitting) return;
		submitting = true;
		try {
			const task = await api.post<Task>(`/repos/${selectedRepoId}/tasks`, { title, status });
			tasks.addTask(task);
			newTaskTitle = '';
			addingToColumn = null;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Failed to create task.';
		} finally {
			submitting = false;
		}
	}

	async function moveTask(taskId: string, newStatus: TaskStatus) {
		const task = tasks.tasks.find((t) => t.id === taskId);
		if (!task || task.status === newStatus || !selectedRepoId) return;
		// Optimistic update
		tasks.updateTask({ ...task, status: newStatus });
		try {
			const updated = await api.patch<Task>(`/repos/${selectedRepoId}/tasks/${taskId}`, { status: newStatus });
			tasks.updateTask(updated);
		} catch (err) {
			// Rollback
			tasks.updateTask(task);
			error = err instanceof ApiError ? err.message : 'Failed to move task.';
		}
	}

	// Drag and drop
	function onDragStart(task: Task) {
		draggingTask = task;
	}

	function onDragOver(e: DragEvent, status: TaskStatus) {
		e.preventDefault();
		dragOverColumn = status;
	}

	function onDragLeave() {
		dragOverColumn = null;
	}

	async function onDrop(e: DragEvent, status: TaskStatus) {
		e.preventDefault();
		dragOverColumn = null;
		if (draggingTask && draggingTask.status !== status) {
			await moveTask(draggingTask.id, status);
		}
		draggingTask = null;
	}

	function onDragEnd() {
		draggingTask = null;
		dragOverColumn = null;
	}

	function priorityLabel(p: number): string {
		return (['', 'P1', 'P2', 'P3', 'P4', 'P5'] as string[])[p] ?? 'P3';
	}

	function priorityColor(p: number): string {
		if (p === 1) return 'text-red-400 border-red-900';
		if (p === 2) return 'text-orange-400 border-orange-900';
		if (p === 4) return 'text-zinc-500 border-zinc-700';
		if (p === 5) return 'text-zinc-600 border-zinc-800';
		return 'text-zinc-400 border-zinc-700';
	}

	function handleNewTaskKeydown(e: KeyboardEvent, status: TaskStatus) {
		if (e.key === 'Enter') createTask(status);
		if (e.key === 'Escape') { addingToColumn = null; newTaskTitle = ''; }
	}

	function selectRepo(id: string) {
		goto(`?repoId=${id}`);
	}
</script>

<div class="flex h-full flex-col">
	<!-- Header -->
	<div class="border-b border-zinc-800 bg-zinc-900 px-4 py-3 sm:px-6 sm:py-4">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<h1 class="text-xl font-bold tracking-tight text-zinc-50">Task Board</h1>
				<p class="mt-0.5 text-sm text-zinc-500">{tasks.tasks.length} task{tasks.tasks.length !== 1 ? 's' : ''}</p>
			</div>

			<!-- Repo selector -->
			{#if repos.length > 0}
				<select
					value={selectedRepoId ?? ''}
					onchange={(e) => selectRepo((e.currentTarget as HTMLSelectElement).value)}
					class="w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 py-1.5 text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500 sm:w-auto"
				>
					<option value="" disabled>Select repository…</option>
					{#each repos as repo (repo.id)}
						<option value={repo.id}>{repo.name}</option>
					{/each}
				</select>
			{/if}

			{#if tasks.loading}
				<Loader2 class="h-4 w-4 animate-spin text-zinc-500" />
			{/if}
		</div>

		{#if error}
			<div class="mt-3 flex items-center gap-2 rounded-md border border-red-900 bg-red-950/50 px-3 py-2 text-xs text-red-400">
				<AlertCircle class="h-3.5 w-3.5 shrink-0" />
				{error}
				<button onclick={() => (error = null)} class="ml-auto text-red-600 hover:text-red-400">✕</button>
			</div>
		{/if}
	</div>

	{#if !selectedRepoId}
		<div class="flex flex-1 items-center justify-center">
			<p class="text-sm text-zinc-500">Select a repository above to view its tasks.</p>
		</div>
	{:else}
		<!-- Kanban board -->
		<div class="flex flex-1 gap-4 overflow-x-auto p-4">
			{#each COLUMNS as col (col.status)}
				<div
					class="flex w-72 shrink-0 flex-col rounded-xl border transition-colors {dragOverColumn === col.status ? 'border-zinc-500 bg-zinc-800/60' : 'border-zinc-800 bg-zinc-900/50'}"
					ondragover={(e) => onDragOver(e, col.status)}
					ondragleave={onDragLeave}
					ondrop={(e) => onDrop(e, col.status)}
					role="list"
				>
					<!-- Column header -->
					<div class="flex items-center justify-between border-b border-zinc-800 px-4 py-3">
						<div class="flex items-center gap-2">
							<span class="text-sm font-semibold {col.color}">{col.label}</span>
							<Badge variant="secondary" class="h-5 min-w-5 justify-center px-1.5 text-xs">
								{byStatus[col.status].length}
							</Badge>
						</div>
						<Button
							variant="ghost"
							size="icon"
							class="h-6 w-6 text-zinc-600 hover:text-zinc-300"
							onclick={() => { addingToColumn = col.status; newTaskTitle = ''; }}
						>
							<Plus class="h-3.5 w-3.5" />
						</Button>
					</div>

					<!-- Task list -->
					<div class="flex flex-1 flex-col gap-2 overflow-y-auto p-3" role="list">
						{#if addingToColumn === col.status}
							<div class="rounded-lg border border-zinc-700 bg-zinc-900 p-3">
								<Input
									bind:value={newTaskTitle}
									placeholder="Task title…"
									class="mb-2 h-7 text-sm"
									onkeydown={(e: KeyboardEvent) => handleNewTaskKeydown(e, col.status)}
								/>
								<div class="flex gap-2">
									<Button
										size="sm"
										class="h-6 text-xs"
										disabled={!newTaskTitle.trim() || submitting}
										onclick={() => createTask(col.status)}
									>
										{submitting ? 'Adding…' : 'Add'}
									</Button>
									<Button
										variant="ghost"
										size="sm"
										class="h-6 text-xs text-zinc-500"
										onclick={() => { addingToColumn = null; newTaskTitle = ''; }}
									>
										Cancel
									</Button>
								</div>
							</div>
						{/if}

						{#each byStatus[col.status] as task (task.id)}
							<div
								draggable="true"
								ondragstart={() => onDragStart(task)}
								ondragend={onDragEnd}
								class="group cursor-grab rounded-lg border border-zinc-800 bg-zinc-900 p-3 transition-all hover:border-zinc-700 active:cursor-grabbing {draggingTask?.id === task.id ? 'opacity-40' : ''}"
								role="listitem"
							>
								<div class="flex items-start gap-2">
									<GripVertical class="mt-0.5 h-3.5 w-3.5 shrink-0 text-zinc-700 opacity-0 transition-opacity group-hover:opacity-100" />
									<div class="min-w-0 flex-1">
										<p class="text-sm font-medium leading-snug text-zinc-100">{task.title}</p>
										{#if task.description}
											<p class="mt-1 line-clamp-2 text-xs text-zinc-500">{task.description}</p>
										{/if}
										<div class="mt-2 flex items-center gap-2">
											<Badge
												variant="outline"
												class="h-4 px-1 text-[10px] {priorityColor(task.priority)}"
											>
												{priorityLabel(task.priority)}
											</Badge>
											{#if task.agent_id}
												<span class="truncate text-[10px] text-zinc-600">
													agent assigned
												</span>
											{/if}
										</div>
									</div>
								</div>
							</div>
						{/each}

						{#if byStatus[col.status].length === 0 && addingToColumn !== col.status}
							<div class="flex flex-1 items-center justify-center py-8">
								<p class="text-xs text-zinc-700">No tasks</p>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
