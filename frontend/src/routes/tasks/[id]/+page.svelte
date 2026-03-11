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

	function senderIcon(type: string): string {
		switch (type) {
			case 'user': return '👤';
			case 'tars': return '🤖';
			case 'system': return '⚙️';
			default: return '💬';
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

	function statusBadgeClass(status: string): string {
		switch (status) {
			case 'open': return 'text-success border-success/30 bg-success/10';
			case 'running': return 'text-success border-success/30 bg-success/10';
			case 'completed': return 'text-warning border-warning/30 bg-warning/10';
			case 'failed': return 'text-danger border-danger/30 bg-danger/10';
			default: return 'text-text-secondary border-border bg-bg-tertiary';
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
		<p class="text-text-secondary font-mono text-sm">Task not found</p>
	</div>
{:else}
	<!-- Header -->
	<header class="flex items-center gap-3 px-6 py-4 border-b border-border shrink-0">
		<h2 class="text-base font-medium text-text-primary truncate">{task.title}</h2>
		<span class="px-2 py-0.5 text-xs font-mono border rounded shrink-0 {statusBadgeClass(task.status)}">
			{task.status}
		</span>
	</header>

	<!-- Messages / Timeline -->
	<div
		bind:this={messagesContainer}
		class="flex-1 overflow-y-auto px-6 py-4 space-y-4"
	>
		{#if messagesStore.loading}
			<div class="flex justify-center py-8">
				<p class="text-text-secondary font-mono text-sm">Loading messages...</p>
			</div>
		{:else if messagesStore.timeline.length === 0}
			<div class="flex flex-col items-center justify-center py-16 text-center">
				<p class="text-text-secondary text-sm">No messages yet. Send a message to start.</p>
			</div>
		{:else}
			{#each messagesStore.timeline as entry (entry.id)}
				{#if isWorkerEvent(entry) && entry.event === 'start'}
					<!-- Worker Card inline in the timeline -->
					<WorkerCard session={getWorkerSession(entry.session_id)} />
				{:else if !isWorkerEvent(entry)}
					{@const message = entry as Message}
					<div class="flex gap-3 {message.sender_type === 'system' ? 'opacity-60' : ''}">
						<!-- Avatar -->
						<div class="w-8 h-8 rounded bg-bg-tertiary flex items-center justify-center shrink-0 text-sm">
							{senderIcon(message.sender_type)}
						</div>

						<!-- Content -->
						<div class="flex-1 min-w-0">
							<div class="flex items-baseline gap-2 mb-1">
								<span class="text-sm font-medium {message.sender_type === 'tars' ? 'text-accent' : 'text-text-primary'}">
									{senderLabel(message.sender_type)}
								</span>
								<span class="text-xs text-text-secondary font-mono">
									{formatTime(message.created_at)}
								</span>
							</div>
							<div class="text-sm text-text-primary leading-relaxed whitespace-pre-wrap">
								{message.content}
							</div>
						</div>
					</div>
				{/if}
			{/each}
		{/if}
	</div>

	<!-- Message input -->
	<form onsubmit={handleSend} class="shrink-0 border-t border-border px-6 py-4">
		<div class="flex gap-3">
			<input
				type="text"
				bind:value={messageInput}
				onkeydown={handleKeydown}
				placeholder="Send a message..."
				class="flex-1 bg-bg-tertiary border border-border rounded-md px-4 py-2.5 text-sm text-text-primary
					placeholder:text-text-secondary focus:outline-none focus:border-accent transition-colors"
			/>
			<button
				type="submit"
				disabled={!messageInput.trim()}
				class="px-5 py-2.5 bg-accent text-bg-primary text-sm font-medium rounded-md
					hover:bg-accent-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
			>
				Send
			</button>
		</div>
	</form>
{/if}
