<script lang="ts">
	import { page } from '$app/state';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { messagesStore, mockWorkerSession } from '$lib/stores/messages.svelte';
	import type { WorkerEvent } from '$lib/stores/messages.svelte';
	import { wsStore } from '$lib/stores/websocket.svelte';
	import { workersStore } from '$lib/stores/workers.svelte';
	import type { Message, WorkerSession } from '$lib/types';
	import WorkerCard from '$lib/components/WorkerCard.svelte';
	import { onMount, tick } from 'svelte';

	let messageInput = $state('');
	let messagesContainer: HTMLDivElement | undefined = $state();

	const taskId = $derived(page.params.id);
	const task = $derived(tasksStore.tasks.find((t) => t.id === taskId) ?? null);

	// Track previous taskId for cleanup
	let prevTaskId = $state<string | null>(null);

	$effect(() => {
		if (taskId) {
			// Unsubscribe from previous task
			if (prevTaskId && prevTaskId !== taskId) {
				wsStore.unsubscribeFromTask(prevTaskId);
			}
			prevTaskId = taskId;

			tasksStore.selectTask(taskId);
			messagesStore.fetchMessages(taskId);
			wsStore.subscribeToTask(taskId);
		}
	});

	// Auto-scroll to bottom when timeline changes
	$effect(() => {
		if (messagesStore.timeline.length && messagesContainer) {
			tick().then(() => {
				if (messagesContainer) {
					messagesContainer.scrollTop = messagesContainer.scrollHeight;
				}
			});
		}
	});

	function isWorkerEvent(entry: unknown): entry is WorkerEvent {
		return (entry as WorkerEvent).type === 'worker_event';
	}

	function getWorkerSession(sessionId: string): WorkerSession {
		const session = workersStore.getWorker(sessionId);
		if (session) return session;
		// Fallback to mock worker for mock data
		if (sessionId === mockWorkerSession.id) return mockWorkerSession;
		// Return a placeholder
		return {
			id: sessionId,
			task_id: taskId,
			status: 'running',
			command: 'claude-code',
			started_at: new Date().toISOString()
		};
	}

	function senderInitial(type: string): string {
		switch (type) {
			case 'user': return 'U';
			case 'tars': return 'T';
			case 'system': return 'S';
			default: return '?';
		}
	}

	function senderLabel(type: string): string {
		switch (type) {
			case 'user': return 'You';
			case 'tars': return 'TARS';
			case 'system': return 'System';
			default: return type;
		}
	}

	function senderAvatarClass(type: string): string {
		switch (type) {
			case 'tars': return 'bg-indigo-500/15 text-indigo-400';
			case 'user': return 'bg-zinc-800 text-zinc-300';
			default: return 'bg-zinc-800/50 text-zinc-500';
		}
	}

	function statusBadgeClass(status: string): string {
		switch (status) {
			case 'open': return 'text-zinc-400 border-zinc-700 bg-zinc-800/50';
			case 'running': return 'text-indigo-400 border-indigo-500/30 bg-indigo-500/10';
			case 'completed': return 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10';
			case 'failed': return 'text-red-400 border-red-500/30 bg-red-500/10';
			default: return 'text-zinc-400 border-zinc-700 bg-zinc-800/50';
		}
	}

	function formatTime(iso: string): string {
		try {
			const d = new Date(iso);
			return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
		} catch {
			return '';
		}
	}

	async function handleSend(e: Event) {
		e.preventDefault();
		if (!messageInput.trim() || !taskId) return;
		const content = messageInput.trim();
		messageInput = '';
		await messagesStore.sendMessage(taskId, content);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSend(e);
		}
	}
</script>

{#if !task}
	<div class="flex-1 flex items-center justify-center">
		<p class="text-zinc-500 text-sm">Task not found</p>
	</div>
{:else}
	<!-- Header -->
	<header class="flex items-center gap-3 px-6 py-4 border-b border-zinc-800/50 shrink-0">
		<h2 class="text-base font-medium text-zinc-100 truncate">{task.title}</h2>
		<span class="px-3 py-1 text-xs font-medium border rounded-full shrink-0 {statusBadgeClass(task.status)}">
			{task.status}
		</span>
	</header>

	<!-- Messages / Timeline -->
	<div
		bind:this={messagesContainer}
		class="flex-1 overflow-y-auto px-6 py-5"
	>
		{#if messagesStore.loading}
			<div class="flex justify-center py-12">
				<p class="text-zinc-500 text-sm">Loading messages...</p>
			</div>
		{:else if messagesStore.timeline.length === 0}
			<div class="flex flex-col items-center justify-center py-20 text-center">
				<p class="text-zinc-500 text-sm">No messages yet. Send a message to start.</p>
			</div>
		{:else}
			<div class="space-y-6">
				{#each messagesStore.timeline as entry (entry.id)}
					{#if isWorkerEvent(entry) && entry.event === 'start'}
						<!-- Worker Card inline in the timeline -->
						<WorkerCard session={getWorkerSession(entry.session_id)} />
					{:else if !isWorkerEvent(entry)}
						{@const message = entry as Message}
						{#if message.sender_type === 'system'}
							<div class="py-2 border-l-2 border-zinc-800 pl-4">
								<p class="text-xs italic text-zinc-500">{message.content}</p>
							</div>
						{:else}
							<div class="flex gap-3.5">
								<!-- Avatar -->
								<div class="w-8 h-8 rounded-full flex items-center justify-center shrink-0 text-sm font-medium {senderAvatarClass(message.sender_type)}">
									{senderInitial(message.sender_type)}
								</div>

								<!-- Content -->
								<div class="flex-1 min-w-0">
									<div class="flex items-baseline gap-2 mb-1">
										<span class="text-sm font-medium text-zinc-300">
											{senderLabel(message.sender_type)}
										</span>
										<span class="text-xs text-zinc-500">
											{formatTime(message.created_at)}
										</span>
									</div>
									<div class="text-sm text-zinc-300 leading-relaxed whitespace-pre-wrap">
										{message.content}
									</div>
								</div>
							</div>
						{/if}
					{/if}
				{/each}
			</div>
		{/if}
	</div>

	<!-- Message input -->
	<form onsubmit={handleSend} class="shrink-0 border-t border-zinc-800/50 px-6 py-4">
		<div class="flex gap-3">
			<input
				type="text"
				bind:value={messageInput}
				onkeydown={handleKeydown}
				placeholder="Send a message..."
				class="flex-1 h-10 bg-zinc-900 border border-zinc-800 rounded-lg px-4 text-sm text-text-primary
					placeholder:text-zinc-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
			/>
			<button
				type="submit"
				disabled={!messageInput.trim()}
				class="h-10 px-5 bg-indigo-500 text-white text-sm font-medium rounded-lg
					hover:bg-indigo-600 disabled:opacity-30 disabled:cursor-not-allowed transition-colors duration-150"
			>
				Send
			</button>
		</div>
	</form>
{/if}
